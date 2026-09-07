package usecaseembedded

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

func (u *usecase) GetConfig(ctx context.Context) (*Config, error) {
	if u.resolver == nil {
		return nil, fmt.Errorf("el módulo de integraciones no está disponible")
	}

	creds, err := u.resolver.GetCachedPlatformCredentials(ctx, whatsAppTypeID)
	if err != nil {
		return nil, fmt.Errorf("credenciales de plataforma no disponibles: %w", err)
	}

	config := &Config{
		Enabled:      boolValue(creds["embedded_signup_enabled"]),
		AppID:        stringValue(creds["app_id"]),
		ConfigID:     stringValue(creds["embedded_signup_config_id"]),
		GraphVersion: graphVersion(stringValue(creds["whatsapp_url"])),
	}

	if config.AppID == "" || config.ConfigID == "" {
		config.Enabled = false
	}

	return config, nil
}

func (u *usecase) Complete(ctx context.Context, businessID uint, input SignupInput) (*SignupResult, error) {
	config, err := u.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if !config.Enabled {
		return nil, fmt.Errorf("el registro insertado no está habilitado")
	}

	if businessID == 0 {
		return nil, fmt.Errorf("business_id es requerido")
	}

	wabaID := strings.TrimSpace(input.WABAID)
	phoneNumberID := strings.TrimSpace(input.PhoneNumberID)
	if wabaID == "" || phoneNumberID == "" {
		return nil, fmt.Errorf("Meta no devolvió el WABA o el número: no se puede completar la conexión")
	}

	integrationID, err := u.resolver.GetIntegrationIDByBusinessAndType(ctx, businessID, whatsAppTypeID)
	if err != nil || integrationID == 0 {
		return nil, fmt.Errorf("el negocio %d no tiene integración de WhatsApp", businessID)
	}

	if err := u.ensureNumberNotTaken(ctx, phoneNumberID, integrationID); err != nil {
		return nil, err
	}

	creds, err := u.resolver.GetCachedPlatformCredentials(ctx, whatsAppTypeID)
	if err != nil {
		return nil, fmt.Errorf("credenciales de plataforma no disponibles: %w", err)
	}

	platform, err := u.credentialsCache.GetWhatsAppDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("credenciales de plataforma no disponibles: %w", err)
	}

	signupAPI := u.signupFactory(platform.WhatsAppURL)

	businessToken, err := signupAPI.ExchangeCode(ctx, config.AppID, stringValue(creds["app_secret"]), input.Code)
	if err != nil {
		return nil, fmt.Errorf("no se pudo canjear el código con Meta: %w", err)
	}

	numbers, err := signupAPI.ListPhoneNumbers(ctx, wabaID, businessToken)
	if err != nil {
		return nil, fmt.Errorf("el token entregado por Meta no permite leer el WABA %s: %w", wabaID, err)
	}

	var encontrado *numeroConectado
	for _, number := range numbers {
		if number.ID == phoneNumberID {
			encontrado = &numeroConectado{
				DisplayPhoneNumber: number.DisplayPhoneNumber,
				VerifiedName:       number.VerifiedName,
				QualityRating:      number.QualityRating,
			}
			break
		}
	}
	if encontrado == nil {
		return nil, fmt.Errorf("el número %s no pertenece al WABA %s según Meta", phoneNumberID, wabaID)
	}

	if err := signupAPI.SubscribeApp(ctx, wabaID, businessToken); err != nil {
		return nil, fmt.Errorf("no se pudo suscribir el webhook al WABA del negocio: %w", err)
	}

	result := &SignupResult{
		IntegrationID:      integrationID,
		BusinessID:         businessID,
		WABAID:             wabaID,
		PhoneNumberID:      phoneNumberID,
		DisplayPhoneNumber: encontrado.DisplayPhoneNumber,
		VerifiedName:       encontrado.VerifiedName,
		QualityRating:      encontrado.QualityRating,
	}

	pin, err := generarPin()
	if err != nil {
		return nil, fmt.Errorf("no se pudo generar el PIN de dos pasos: %w", err)
	}

	numbersAPI := u.numbersFactory(platform.WhatsAppURL)
	if regErr := numbersAPI.Register(ctx, phoneNumberID, businessToken, pin); regErr != nil {
		result.Warning = fmt.Sprintf("el número quedó conectado pero Meta no lo registró: %v", regErr)
		u.log.Warn(ctx).Err(regErr).
			Uint("business_id", businessID).
			Str("phone_number_id", phoneNumberID).
			Msg("registro insertado: el numero no quedo registrado en la Cloud API")
	} else {
		result.Registered = true
		result.Pin = pin
	}

	credentials := map[string]any{"access_token": businessToken}
	if result.Registered {
		credentials["two_factor_pin"] = pin
	}
	if err := u.resolver.UpdateIntegrationCredentials(ctx, strconv.FormatUint(uint64(integrationID), 10), credentials); err != nil {
		return nil, fmt.Errorf("no se pudo guardar el token del negocio: %w", err)
	}

	numberStatus := "verificado"
	usePlatformToken := true
	if result.Registered {
		numberStatus = "registrado"
		usePlatformToken = false
	}

	if err := u.resolver.UpdateIntegrationConfig(ctx, strconv.FormatUint(uint64(integrationID), 10), map[string]any{
		"waba_id":            wabaID,
		"phone_number_id":    phoneNumberID,
		"hosted_by_platform": false,
		"number_status":      numberStatus,
		"verified_name":      encontrado.VerifiedName,
		"use_platform_token": usePlatformToken,
		"embedded_signup":    true,
	}); err != nil {
		return nil, fmt.Errorf("no se pudo guardar la conexión: %w", err)
	}

	u.log.Info(ctx).
		Uint("business_id", businessID).
		Str("waba_id", wabaID).
		Str("phone_number_id", phoneNumberID).
		Bool("registrado", result.Registered).
		Msg("registro insertado completado")

	return result, nil
}

func (u *usecase) ensureNumberNotTaken(ctx context.Context, phoneNumberID string, integrationID uint) error {
	ownerID, ownerBusinessID, err := u.resolver.FindIntegrationByConfigValue(ctx, whatsAppTypeID, "phone_number_id", phoneNumberID)
	if err != nil {
		return fmt.Errorf("error verificando si el número ya está en uso: %w", err)
	}
	if ownerID == 0 || ownerID == integrationID {
		return nil
	}

	u.log.Warn(ctx).
		Str("phone_number_id", phoneNumberID).
		Uint("owner_integration_id", ownerID).
		Uint("owner_business_id", ownerBusinessID).
		Msg("registro insertado: el numero ya pertenece a otra integracion")

	return fmt.Errorf("ese número ya está conectado a otra cuenta de Probability")
}

type numeroConectado struct {
	DisplayPhoneNumber string
	VerifiedName       string
	QualityRating      string
}

func graphVersion(whatsappURL string) string {
	for _, parte := range strings.Split(whatsappURL, "/") {
		if strings.HasPrefix(parte, "v") && strings.Contains(parte, ".") {
			return parte
		}
	}
	return "v22.0"
}

func boolValue(raw any) bool {
	switch v := raw.(type) {
	case bool:
		return v
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		return err == nil && parsed
	}
	return false
}

func stringValue(raw any) string {
	value, _ := raw.(string)
	return strings.TrimSpace(value)
}

func generarPin() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

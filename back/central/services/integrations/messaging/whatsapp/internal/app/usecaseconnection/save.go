package usecaseconnection

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

func (u *usecase) SaveConnection(ctx context.Context, businessID uint, input SaveConnectionInput) (*ConnectionResult, error) {
	if u.resolver == nil {
		return nil, fmt.Errorf("el módulo de integraciones no está disponible")
	}
	if businessID == 0 {
		return nil, fmt.Errorf("business_id es requerido")
	}

	integrationID, err := u.resolver.GetIntegrationIDByBusinessAndType(ctx, businessID, whatsAppTypeID)
	if err != nil || integrationID == 0 {
		return nil, fmt.Errorf("el negocio %d no tiene integración de WhatsApp", businessID)
	}

	if input.UsePlatformToken {
		return u.disconnectOwnNumber(ctx, integrationID, businessID)
	}

	wabaID := strings.TrimSpace(input.WABAID)
	phoneNumberID := strings.TrimSpace(input.PhoneNumberID)
	accessToken := strings.TrimSpace(input.AccessToken)

	if phoneNumberID == "" {
		return nil, fmt.Errorf("se necesita el phone_number_id del número")
	}
	if _, err := strconv.ParseUint(phoneNumberID, 10, 64); err != nil {
		return nil, fmt.Errorf("phone_number_id inválido: %s", phoneNumberID)
	}

	platform, err := u.credentialsCache.GetWhatsAppDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("credenciales de plataforma no disponibles: %w", err)
	}

	hosted := wabaID == ""
	if hosted {
		if platform.WABAID == "" {
			return nil, fmt.Errorf("las credenciales de plataforma no tienen waba_id: no se puede alojar el número en la cuenta de Probability")
		}
		wabaID = platform.WABAID
	}
	if _, err := strconv.ParseUint(wabaID, 10, 64); err != nil {
		return nil, fmt.Errorf("waba_id inválido: %s", wabaID)
	}

	if platform.PhoneNumberID != 0 && phoneNumberID == strconv.FormatUint(uint64(platform.PhoneNumberID), 10) {
		return nil, fmt.Errorf("ese phone_number_id es el número de Probability: usa la opción de número de plataforma")
	}

	if err := u.ensureNumberNotTaken(ctx, phoneNumberID, integrationID); err != nil {
		return nil, err
	}

	effectiveToken := accessToken
	platformToken := false
	if effectiveToken == "" {
		if stored := u.storedAccessToken(ctx, integrationID); stored != "" {
			effectiveToken = stored
		} else {
			effectiveToken = platform.AccessToken
			platformToken = true
		}
	}
	if effectiveToken == "" {
		return nil, fmt.Errorf("no hay token disponible para validar el número contra Meta")
	}

	number, err := u.verifyAgainstMeta(ctx, platform.WhatsAppURL, wabaID, phoneNumberID, effectiveToken)
	if err != nil {
		return nil, err
	}

	config := map[string]any{
		"use_platform_token": false,
		"waba_id":            wabaID,
		"phone_number_id":    phoneNumberID,
		"hosted_by_platform": hosted,
	}
	if err := u.resolver.UpdateIntegrationConfig(ctx, strconv.FormatUint(uint64(integrationID), 10), config); err != nil {
		return nil, fmt.Errorf("error guardando la configuración de la conexión: %w", err)
	}

	if accessToken != "" {
		credentials := map[string]any{"access_token": accessToken}
		if err := u.resolver.UpdateIntegrationCredentials(ctx, strconv.FormatUint(uint64(integrationID), 10), credentials); err != nil {
			return nil, fmt.Errorf("error guardando el token de la conexión: %w", err)
		}
	}

	u.log.Info(ctx).
		Uint("business_id", businessID).
		Uint("integration_id", integrationID).
		Str("waba_id", wabaID).
		Str("phone_number_id", phoneNumberID).
		Bool("token_de_plataforma", platformToken).
		Bool("alojado_en_probability", hosted).
		Msg("número propio de WhatsApp conectado y verificado contra Meta")

	return &ConnectionResult{
		IntegrationID:      integrationID,
		BusinessID:         businessID,
		OwnNumber:          true,
		WABAID:             wabaID,
		PhoneNumberID:      phoneNumberID,
		DisplayPhoneNumber: number.DisplayPhoneNumber,
		VerifiedName:       number.VerifiedName,
		QualityRating:      number.QualityRating,
		PlatformToken:      platformToken,
		HostedByPlatform:   hosted,
	}, nil
}

func (u *usecase) disconnectOwnNumber(ctx context.Context, integrationID, businessID uint) (*ConnectionResult, error) {
	config := map[string]any{
		"use_platform_token": true,
		"waba_id":            "",
		"phone_number_id":    "",
		"hosted_by_platform": false,
	}
	if err := u.resolver.UpdateIntegrationConfig(ctx, strconv.FormatUint(uint64(integrationID), 10), config); err != nil {
		return nil, fmt.Errorf("error volviendo al número de la plataforma: %w", err)
	}

	u.log.Info(ctx).
		Uint("business_id", businessID).
		Uint("integration_id", integrationID).
		Msg("el negocio vuelve al número de WhatsApp de la plataforma")

	return &ConnectionResult{
		IntegrationID: integrationID,
		BusinessID:    businessID,
		OwnNumber:     false,
	}, nil
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
		Uint("integration_id", integrationID).
		Uint("owner_integration_id", ownerID).
		Uint("owner_business_id", ownerBusinessID).
		Msg("intento de conectar un phone_number_id que ya pertenece a otra integración")

	return fmt.Errorf("ese phone_number_id ya está conectado a otra cuenta")
}

func (u *usecase) storedAccessToken(ctx context.Context, integrationID uint) string {
	_, credentials, err := u.resolver.GetIntegrationConfigAndCredentials(ctx, integrationID)
	if err != nil {
		return ""
	}
	token, _ := credentials["access_token"].(string)
	return strings.TrimSpace(token)
}

func (u *usecase) verifyAgainstMeta(ctx context.Context, baseURL, wabaID, phoneNumberID, accessToken string) (*metaNumber, error) {
	api := u.apiFactory(baseURL)

	numbers, err := api.ListPhoneNumbers(ctx, wabaID, accessToken)
	if err != nil {
		return nil, fmt.Errorf("Meta no permitió leer los números del WABA %s: %w", wabaID, err)
	}

	for _, number := range numbers {
		if number.ID == phoneNumberID {
			return &metaNumber{
				DisplayPhoneNumber: number.DisplayPhoneNumber,
				VerifiedName:       number.VerifiedName,
				QualityRating:      number.QualityRating,
			}, nil
		}
	}

	return nil, fmt.Errorf("el phone_number_id %s no pertenece al WABA %s según Meta", phoneNumberID, wabaID)
}

type metaNumber struct {
	DisplayPhoneNumber string
	VerifiedName       string
	QualityRating      string
}

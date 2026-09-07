package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/secamc93/probability/back/central/services/auth/demo/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (uc *UseCase) DemoRegisterWithGoogle(ctx context.Context, request domain.GoogleDemoRegisterRequest) (*domain.GoogleDemoRegisterResponse, error) {
	if uc.signupTokens == nil || uc.sessionTokens == nil {
		return nil, fmt.Errorf("el registro con Google no esta disponible")
	}

	businessName := strings.TrimSpace(request.BusinessName)
	if businessName == "" {
		return nil, domain.ErrBusinessNameRequired
	}

	claims, err := uc.signupTokens.ValidateGoogleSignupToken(strings.TrimSpace(request.SignupToken))
	if err != nil {
		uc.log.Warn().Err(err).Msg("Token de registro con Google invalido o expirado")
		return nil, domain.ErrGoogleSignupTokenInvalid
	}

	email := strings.ToLower(strings.TrimSpace(claims.Email))
	fullName := strings.TrimSpace(claims.Name)
	if fullName == "" {
		fullName = email
	}

	existing, err := uc.repository.GetDemoUserByEmail(ctx, email)
	if err != nil {
		uc.log.Error().Err(err).Msg("Error verificando email en registro demo con Google")
		return nil, fmt.Errorf("error interno del servidor")
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyRegistered
	}

	roleID, err := uc.repository.GetDemoRoleID(ctx)
	if err != nil || roleID == 0 {
		uc.log.Error().Err(err).Msg("Rol demo no encontrado")
		return nil, fmt.Errorf("error interno del servidor")
	}

	businessCode, err := uc.uniqueBusinessCode(ctx, businessName)
	if err != nil {
		uc.log.Error().Err(err).Msg("Error generando codigo de negocio")
		return nil, fmt.Errorf("error interno del servidor")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(randomPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error interno del servidor")
	}

	cuenta, err := uc.repository.CreateDemoAccount(ctx, domain.CreateDemoAccountParams{
		GoogleID:     claims.GoogleID,
		Active:       true,
		AvatarURL:    claims.Picture,
		FullName:     fullName,
		BusinessName: businessName,
		BusinessCode: businessCode,
		OrderPrefix:  derivePrefix(businessName),
		Email:        email,
		PasswordHash: string(hashed),
		RoleID:       roleID,
		ExpiresAt:    time.Now(),
	})
	if err != nil {
		uc.log.Error().Err(err).Str("email", email).Msg("Error creando cuenta demo con Google")
		return nil, fmt.Errorf("no se pudo crear la cuenta demo")
	}

	if uc.onBusinessCreated != nil {
		uc.onBusinessCreated(ctx, cuenta.BusinessID)
	}

	if err := uc.repository.ProvisionDemoIntegrations(ctx, cuenta.BusinessID, cuenta.UserID); err != nil {
		uc.log.Warn().Err(err).Uint("business_id", cuenta.BusinessID).Msg("No se pudieron aprovisionar las integraciones demo")
	}

	token, err := uc.sessionTokens.GenerateToken(cuenta.UserID, cuenta.BusinessID, 1, cuenta.RoleID, "active")
	if err != nil {
		uc.log.Error().Err(err).Uint("user_id", cuenta.UserID).Msg("Error generando el token de sesion de la cuenta demo")
		return nil, fmt.Errorf("error interno del servidor")
	}

	uc.log.Info().
		Str("email", email).
		Str("business", businessName).
		Uint("business_id", cuenta.BusinessID).
		Msg("Cuenta demo creada con Google")

	return &domain.GoogleDemoRegisterResponse{
		Success:      true,
		Message:      "Cuenta demo creada",
		Token:        token,
		UserID:       cuenta.UserID,
		BusinessID:   cuenta.BusinessID,
		FullName:     fullName,
		Email:        email,
		BusinessName: businessName,
		AvatarURL:    claims.Picture,
	}, nil
}

func randomPassword() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return time.Now().Format(time.RFC3339Nano)
	}
	return hex.EncodeToString(b)
}

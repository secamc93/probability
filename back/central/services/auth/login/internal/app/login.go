package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

func (uc *AuthUseCase) Login(ctx context.Context, request domain.LoginRequest) (*domain.LoginResponse, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(request.Email))
	uc.log.Info().Str("email", normalizedEmail).Msg("Iniciando proceso de login")

	if normalizedEmail == "" || request.Password == "" {
		uc.log.Error().Msg("Email o contraseña vacíos")
		return nil, domain.ErrEmailPasswordRequired
	}

	userAuth, err := uc.repository.GetUserByEmail(ctx, normalizedEmail)
	if err != nil {
		uc.log.Error().Err(err).Str("email", request.Email).Msg("Error al obtener usuario por email")
		return nil, fmt.Errorf("error al obtener usuario por email: %w", err)
	}

	if userAuth == nil {
		uc.log.Error().Str("email", normalizedEmail).Msg("Usuario no encontrado")
		return nil, domain.ErrUserNotFound
	}

	uc.log.Debug().
		Str("email", request.Email).
		Msg("Validando contraseña con bcrypt")

	if err := bcrypt.CompareHashAndPassword([]byte(userAuth.Password), []byte(request.Password)); err != nil {
		uc.log.Error().
			Err(err).
			Str("email", request.Email).
			Msg("Contraseña inválida")
		return nil, domain.ErrInvalidCredentials
	}

	if !userAuth.IsActive {
		pending, err := uc.repository.HasPendingEmailVerification(ctx, userAuth.ID)
		if err != nil {
			uc.log.Error().Err(err).Uint("user_id", userAuth.ID).Msg("Error consultando verificación pendiente")
			return nil, domain.ErrUserInactive
		}
		if pending {
			uc.log.Warn().Str("email", request.Email).Msg("Usuario inactivo pendiente de verificación")
			return nil, domain.ErrUserPendingVerification
		}
		uc.log.Error().Str("email", request.Email).Msg("Usuario inactivo")
		return nil, domain.ErrUserInactive
	}

	return uc.buildSession(ctx, userAuth)
}

func isSuperAdmin(roles []domain.Role) bool {
	for _, role := range roles {
		if role.ScopeCode == "platform" {
			return true
		}
	}
	return false
}

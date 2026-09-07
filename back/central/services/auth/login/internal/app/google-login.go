package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
)

func (uc *AuthUseCase) GoogleAuthURL(ctx context.Context) (*domain.GoogleAuthURLResponse, error) {
	if uc.googleOAuth == nil || !uc.googleOAuth.IsConfigured() {
		uc.log.Error().Msg("Google OAuth no esta configurado")
		return nil, domain.ErrGoogleNotConfigured
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("error generando state de Google: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(raw)

	authURL, err := uc.googleOAuth.AuthCodeURL(state)
	if err != nil {
		return nil, err
	}

	return &domain.GoogleAuthURLResponse{AuthURL: authURL, State: state}, nil
}

func (uc *AuthUseCase) LoginWithGoogle(ctx context.Context, request domain.GoogleCallbackRequest) (*domain.GoogleLoginResult, error) {
	if uc.googleOAuth == nil || !uc.googleOAuth.IsConfigured() {
		uc.log.Error().Msg("Google OAuth no esta configurado")
		return nil, domain.ErrGoogleNotConfigured
	}

	if strings.TrimSpace(request.Code) == "" {
		return nil, domain.ErrGoogleCodeRequired
	}

	profile, err := uc.googleOAuth.ExchangeCode(ctx, request.Code)
	if err != nil {
		return nil, err
	}

	if !profile.EmailVerified {
		uc.log.Warn().Str("email", profile.Email).Msg("Cuenta de Google sin correo verificado")
		return nil, domain.ErrGoogleEmailNotVerified
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(profile.Email))

	userAuth, err := uc.repository.GetUserByGoogleID(ctx, profile.Sub)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario por google_id: %w", err)
	}

	if userAuth == nil {
		userAuth, err = uc.repository.GetUserByEmail(ctx, normalizedEmail)
		if err != nil {
			return nil, fmt.Errorf("error al obtener usuario por email: %w", err)
		}
		if userAuth == nil {
			uc.log.Warn().Str("email", normalizedEmail).Msg("Login con Google de un correo sin cuenta")
			return &domain.GoogleLoginResult{Profile: profile, NeedsSignup: true}, nil
		}
		if err := uc.repository.LinkGoogleAccount(ctx, userAuth.ID, profile.Sub); err != nil {
			return nil, domain.ErrGoogleAccountLinkedElsewhere
		}
		uc.log.Info().
			Uint("user_id", userAuth.ID).
			Str("email", normalizedEmail).
			Msg("Cuenta de Google vinculada al usuario existente")
	}

	if !userAuth.IsActive {
		pending, err := uc.repository.HasPendingEmailVerification(ctx, userAuth.ID)
		if err != nil {
			return nil, domain.ErrUserInactive
		}
		if pending {
			return nil, domain.ErrUserPendingVerification
		}
		return nil, domain.ErrUserInactive
	}

	if userAuth.AvatarURL == "" && profile.Picture != "" {
		if err := uc.repository.UpdateAvatarIfEmpty(ctx, userAuth.ID, profile.Picture); err == nil {
			userAuth.AvatarURL = profile.Picture
		}
	}

	uc.log.Info().
		Uint("user_id", userAuth.ID).
		Str("email", normalizedEmail).
		Msg("Login con Google validado")

	session, err := uc.buildSession(ctx, userAuth)
	if err != nil {
		return nil, err
	}

	return &domain.GoogleLoginResult{Session: session, Profile: profile}, nil
}

func (uc *AuthUseCase) GoogleSignupToken(ctx context.Context, profile *domain.GoogleProfile, ttl time.Duration) (string, error) {
	if profile == nil {
		return "", domain.ErrGoogleExchangeFailed
	}
	return uc.jwtService.GenerateGoogleSignupToken(profile.Sub, profile.Email, profile.Name, profile.Picture, ttl)
}

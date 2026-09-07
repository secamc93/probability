package google

import (
	"context"
	"errors"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/shared/googleoauth"
)

type adapter struct {
	client googleoauth.IClient
}

func New(client googleoauth.IClient) domain.IGoogleOAuthProvider {
	return &adapter{client: client}
}

func (a *adapter) IsConfigured() bool {
	return a.client.IsConfigured()
}

func (a *adapter) AuthCodeURL(state string) (string, error) {
	url, err := a.client.AuthCodeURL(state)
	return url, traducir(err)
}

func (a *adapter) ExchangeCode(ctx context.Context, code string) (*domain.GoogleProfile, error) {
	perfil, err := a.client.ExchangeCode(ctx, code)
	if err != nil {
		return nil, traducir(err)
	}
	return &domain.GoogleProfile{
		Sub:           perfil.Sub,
		Email:         perfil.Email,
		EmailVerified: perfil.EmailVerified,
		Name:          perfil.Name,
		Picture:       perfil.Picture,
	}, nil
}

func traducir(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, googleoauth.ErrNotConfigured):
		return domain.ErrGoogleNotConfigured
	case errors.Is(err, googleoauth.ErrExchangeFailed):
		return domain.ErrGoogleExchangeFailed
	default:
		return err
	}
}

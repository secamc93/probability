package app

import (
	"context"
	"fmt"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
)

func (uc *AuthUseCase) CurrentSession(ctx context.Context, userID uint) (*domain.LoginResponse, error) {
	userAuth, err := uc.repository.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error al obtener usuario por id: %w", err)
	}
	if userAuth == nil {
		return nil, domain.ErrUserNotFound
	}
	if !userAuth.IsActive {
		return nil, domain.ErrUserInactive
	}
	return uc.buildSession(ctx, userAuth)
}

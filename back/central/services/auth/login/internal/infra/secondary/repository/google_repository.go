package repository

import (
	"context"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/migration/shared/models"
	"gorm.io/gorm"
)

func (r *Repository) GetUserByGoogleID(ctx context.Context, googleID string) (*domain.UserAuthInfo, error) {
	var user domain.UserAuthInfo
	if err := r.database.Conn(ctx).
		Model(&models.User{}).
		Where("google_id = ?", googleID).
		First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.logger.Error().Str("google_id", googleID).Err(err).Msg("Error al obtener usuario por google_id")
		return nil, err
	}
	return &user, nil
}

func (r *Repository) LinkGoogleAccount(ctx context.Context, userID uint, googleID string) error {
	if err := r.database.Conn(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("google_id", googleID).Error; err != nil {
		r.logger.Error().Uint("user_id", userID).Err(err).Msg("Error al vincular cuenta de Google")
		return err
	}
	return nil
}

func (r *Repository) UpdateAvatarIfEmpty(ctx context.Context, userID uint, avatarURL string) error {
	if avatarURL == "" {
		return nil
	}
	if err := r.database.Conn(ctx).
		Model(&models.User{}).
		Where("id = ? AND coalesce(avatar_url, '') = ''", userID).
		Update("avatar_url", avatarURL).Error; err != nil {
		r.logger.Warn().Uint("user_id", userID).Err(err).Msg("Error al actualizar avatar desde Google")
		return err
	}
	return nil
}

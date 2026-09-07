package tokens

import (
	"github.com/secamc93/probability/back/central/services/auth/demo/internal/domain"
	"github.com/secamc93/probability/back/central/shared/jwt"
)

type adapter struct {
	service jwt.IJWTService
}

func New(service jwt.IJWTService) *adapter {
	return &adapter{service: service}
}

func (a *adapter) ValidateGoogleSignupToken(tokenString string) (*domain.GoogleSignupClaims, error) {
	claims, err := a.service.ValidateGoogleSignupToken(tokenString)
	if err != nil {
		return nil, err
	}
	return &domain.GoogleSignupClaims{
		GoogleID: claims.GoogleID,
		Email:    claims.Email,
		Name:     claims.Name,
		Picture:  claims.Picture,
	}, nil
}

func (a *adapter) GenerateToken(userID, businessID, businessTypeID, roleID uint, subscriptionStatus string) (string, error) {
	return a.service.GenerateToken(userID, businessID, businessTypeID, roleID, subscriptionStatus)
}

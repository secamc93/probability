package domain

import (
	"context"
	"time"
)

type IAuthRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*UserAuthInfo, error)
	HasPendingEmailVerification(ctx context.Context, userID uint) (bool, error)
	GetUserByID(ctx context.Context, userID uint) (*UserAuthInfo, error)
	CreatePasswordResetToken(ctx context.Context, userID uint, tokenHash string, channel string, expiresAt time.Time) error
	InvalidateUserPasswordResetTokens(ctx context.Context, userID uint) error
	GetValidPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetTokenInfo, error)
	GetActiveOTPToken(ctx context.Context, userID uint) (*PasswordResetTokenInfo, error)
	IncrementPasswordResetTokenAttempts(ctx context.Context, tokenID uint) error
	MarkPasswordResetTokenUsed(ctx context.Context, tokenID uint) error
	GetUserRoles(ctx context.Context, userID uint) ([]Role, error)
	GetRolePermissions(ctx context.Context, roleID uint) ([]Permission, error)
	UpdateLastLogin(ctx context.Context, userID uint) error
	ChangePassword(ctx context.Context, userID uint, newPassword string) error
	GetUserBusinesses(ctx context.Context, userID uint) ([]BusinessInfoEntity, error)
	GetUserRoleByBusiness(ctx context.Context, userID uint, businessID uint) (*Role, error)
	GetBusinessStaffRelation(ctx context.Context, userID uint, businessID *uint) (*BusinessStaffRelation, error)
	GetBusinessConfiguredResourcesIDs(ctx context.Context, businessID uint) ([]uint, error)
	GetBusinessByID(ctx context.Context, businessID uint) (*BusinessInfo, error)
	GetRoleByID(ctx context.Context, id uint) (*Role, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*UserAuthInfo, error)
	LinkGoogleAccount(ctx context.Context, userID uint, googleID string) error
	UpdateAvatarIfEmpty(ctx context.Context, userID uint, avatarURL string) error
}

type IGoogleOAuthProvider interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*GoogleProfile, error)
	IsConfigured() bool
}
type IJWTService interface {
	GenerateToken(userID, businessID, businessTypeID, roleID uint, subscriptionStatus string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(tokenString string) (string, error)
}

type IEmailSender interface {
	SendHTML(ctx context.Context, to, subject, html string) error
}

type IOTPEventPublisher interface {
	PublishPasswordResetOTP(ctx context.Context, event PasswordResetOTPEvent) error
}

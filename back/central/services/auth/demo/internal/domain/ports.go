package domain

import (
	"context"
	"time"
)

type IDemoRepository interface {
	GetDemoUserByEmail(ctx context.Context, email string) (*PendingDemoUser, error)
	InvalidateEmailVerificationTokens(ctx context.Context, userID uint) error
	CreateEmailVerificationToken(ctx context.Context, userID uint, tokenHash string, expiresAt time.Time) error
	UpdateUserPhone(ctx context.Context, userID uint, phone string) error
	BusinessCodeExists(ctx context.Context, code string) (bool, error)
	GetDemoRoleID(ctx context.Context) (uint, error)
	CreateDemoAccount(ctx context.Context, params CreateDemoAccountParams) (*DemoAccountCreated, error)
	GetValidEmailVerificationToken(ctx context.Context, tokenHash string) (*EmailVerificationTokenInfo, error)
	ActivateUserAndConsumeToken(ctx context.Context, tokenID, userID uint) error
	GetBusinessIDByUserID(ctx context.Context, userID uint) (uint, error)
	ProvisionDemoIntegrations(ctx context.Context, businessID, userID uint) error
}

type IGoogleSignupTokenValidator interface {
	ValidateGoogleSignupToken(tokenString string) (*GoogleSignupClaims, error)
}

type ISessionTokenIssuer interface {
	GenerateToken(userID, businessID, businessTypeID, roleID uint, subscriptionStatus string) (string, error)
}

type IEmailSender interface {
	SendHTML(ctx context.Context, to, subject, html string) error
}

type IDemoOTPPublisher interface {
	PublishDemoOTP(ctx context.Context, event DemoOTPEvent) error
}

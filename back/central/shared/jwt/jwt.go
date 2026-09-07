package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type IJWTService interface {
	GenerateToken(userID, businessID, businessTypeID, roleID uint, subscriptionStatus string) (string, error)
	ValidateToken(tokenString string) (*JWTClaims, error)
	RefreshToken(tokenString string) (string, error)

	GeneratePublicVotingToken(votingID, votingGroupID, hpID uint, durationHours int) (string, error)
	GenerateVotingAuthToken(residentID, propertyUnitID, votingID, votingGroupID, hpID uint) (string, error)
	ValidatePublicVotingToken(tokenString string) (*PublicVotingClaims, error)
	ValidateVotingAuthToken(tokenString string) (*VotingAuthClaims, error)

	GenerateGoogleSignupToken(googleID, email, name, picture string, ttl time.Duration) (string, error)
	ValidateGoogleSignupToken(tokenString string) (*GoogleSignupClaims, error)
}

type JWTService struct {
	secretKey string
}

type Claims struct {
	UserID             uint   `json:"user_id"`
	BusinessID         uint   `json:"business_id"`
	BusinessTypeID     uint   `json:"business_type_id"`
	RoleID             uint   `json:"role_id"`
	SubscriptionStatus string `json:"subscription_status"`
	jwt.RegisteredClaims
}

type JWTClaims struct {
	UserID             uint
	BusinessID         uint
	BusinessTypeID     uint
	RoleID             uint
	SubscriptionStatus string
}

func New(secretKey string) IJWTService {
	return &JWTService{
		secretKey: secretKey,
	}
}

func (j *JWTService) GenerateToken(userID, businessID, businessTypeID, roleID uint, subscriptionStatus string) (string, error) {
	claims := Claims{
		UserID:             userID,
		BusinessID:         businessID,
		BusinessTypeID:     businessTypeID,
		RoleID:             roleID,
		SubscriptionStatus: subscriptionStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "central-reserve-api",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", fmt.Errorf("error al firmar token: %w", err)
	}

	return tokenString, nil
}

func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {

	parser := jwt.NewParser(jwt.WithLeeway(5 * time.Minute))

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error al parsear token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return &JWTClaims{
			UserID:             claims.UserID,
			BusinessID:         claims.BusinessID,
			BusinessTypeID:     claims.BusinessTypeID,
			RoleID:             claims.RoleID,
			SubscriptionStatus: claims.SubscriptionStatus,
		}, nil
	}

	return nil, fmt.Errorf("token inválido")
}

func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	return j.GenerateToken(claims.UserID, claims.BusinessID, claims.BusinessTypeID, claims.RoleID, claims.SubscriptionStatus)
}

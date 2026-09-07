package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const googleSignupPurpose = "google_demo_signup"

type googleSignupClaims struct {
	Purpose  string `json:"purpose"`
	GoogleID string `json:"google_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	jwt.RegisteredClaims
}

type GoogleSignupClaims struct {
	GoogleID string
	Email    string
	Name     string
	Picture  string
}

func (j *JWTService) GenerateGoogleSignupToken(googleID, email, name, picture string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := googleSignupClaims{
		Purpose:  googleSignupPurpose,
		GoogleID: googleID,
		Email:    email,
		Name:     name,
		Picture:  picture,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "central-reserve-api",
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTService) ValidateGoogleSignupToken(tokenString string) (*GoogleSignupClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &googleSignupClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de firma inesperado: %v", t.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*googleSignupClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("token de registro con Google invalido")
	}
	if claims.Purpose != googleSignupPurpose {
		return nil, fmt.Errorf("token de registro con Google invalido")
	}
	if claims.GoogleID == "" || claims.Email == "" {
		return nil, fmt.Errorf("token de registro con Google incompleto")
	}

	return &GoogleSignupClaims{
		GoogleID: claims.GoogleID,
		Email:    claims.Email,
		Name:     claims.Name,
		Picture:  claims.Picture,
	}, nil
}

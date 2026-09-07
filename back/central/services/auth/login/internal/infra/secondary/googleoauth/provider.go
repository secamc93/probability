package googleoauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
)

func (p *Provider) AuthCodeURL(state string) (string, error) {
	if !p.IsConfigured() {
		return "", domain.ErrGoogleNotConfigured
	}

	params := url.Values{}
	params.Set("client_id", p.clientID)
	params.Set("redirect_uri", p.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", scopes)
	params.Set("state", state)
	params.Set("access_type", "online")
	params.Set("prompt", "select_account")

	return fmt.Sprintf("%s?%s", authEndpoint, params.Encode()), nil
}

func (p *Provider) ExchangeCode(ctx context.Context, code string) (*domain.GoogleProfile, error) {
	if !p.IsConfigured() {
		return nil, domain.ErrGoogleNotConfigured
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("redirect_uri", p.redirectURI)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.logger.Error().Err(err).Msg("Error llamando al token endpoint de Google")
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		p.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("Google rechazo el intercambio del codigo")
		return nil, domain.ErrGoogleExchangeFailed
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}
	if tokenResp.IDToken == "" {
		p.logger.Error().Msg("Google no devolvio id_token")
		return nil, domain.ErrGoogleExchangeFailed
	}

	return parseIDToken(tokenResp.IDToken, p.clientID)
}

func parseIDToken(idToken string, clientID string) (*domain.GoogleProfile, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, domain.ErrGoogleExchangeFailed
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}

	var claims struct {
		Iss           string `json:"iss"`
		Aud           string `json:"aud"`
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrGoogleExchangeFailed, err)
	}

	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, domain.ErrGoogleExchangeFailed
	}
	if claims.Aud != clientID {
		return nil, domain.ErrGoogleExchangeFailed
	}
	if claims.Sub == "" || claims.Email == "" {
		return nil, domain.ErrGoogleExchangeFailed
	}

	return &domain.GoogleProfile{
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

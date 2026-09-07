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
)

func (c *client) AuthCodeURL(state string) (string, error) {
	if !c.IsConfigured() {
		return "", ErrNotConfigured
	}

	params := url.Values{}
	params.Set("client_id", c.clientID)
	params.Set("redirect_uri", c.redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", scopes)
	params.Set("state", state)
	params.Set("access_type", "online")
	params.Set("prompt", "select_account")

	return fmt.Sprintf("%s?%s", authEndpoint, params.Encode()), nil
}

func (c *client) ExchangeCode(ctx context.Context, code string) (*Profile, error) {
	if !c.IsConfigured() {
		return nil, ErrNotConfigured
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", c.clientID)
	form.Set("client_secret", c.clientSecret)
	form.Set("redirect_uri", c.redirectURI)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("Error llamando al token endpoint de Google")
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().
			Int("status", resp.StatusCode).
			Str("body", string(body)).
			Msg("Google rechazo el intercambio del codigo")
		return nil, ErrExchangeFailed
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}
	if tokenResp.IDToken == "" {
		c.logger.Error().Msg("Google no devolvio id_token")
		return nil, ErrExchangeFailed
	}

	return parseIDToken(tokenResp.IDToken, c.clientID)
}

func parseIDToken(idToken string, clientID string) (*Profile, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, ErrExchangeFailed
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
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
		return nil, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}

	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, ErrExchangeFailed
	}
	if claims.Aud != clientID {
		return nil, ErrExchangeFailed
	}
	if claims.Sub == "" || claims.Email == "" {
		return nil, ErrExchangeFailed
	}

	return &Profile{
		Sub:           claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
	}, nil
}

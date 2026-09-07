package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/domain/ports"
	"github.com/secamc93/probability/back/central/shared/httpclient"
	"github.com/secamc93/probability/back/central/shared/log"
)

type embeddedSignupClient struct {
	httpClient *httpclient.Client
	templates  ports.ITemplateAPI
	logger     log.ILogger
}

func NewEmbeddedSignupClient(baseURL string, logger log.ILogger) ports.IEmbeddedSignupAPI {
	baseURL = strings.TrimRight(baseURL, "/")

	client := httpclient.New(httpclient.HTTPClientConfig{
		Timeout:    30 * time.Second,
		BaseURL:    baseURL,
		RetryCount: 1,
		RetryWait:  3 * time.Second,
		Debug:      false,
	}, logger.WithModule("whatsapp-embedded-signup-client"))

	client.SetHeader("Content-Type", "application/json")

	return &embeddedSignupClient{
		httpClient: client,
		templates:  NewTemplatesClient(baseURL, logger),
		logger:     logger.WithModule("whatsapp-embedded-signup-client"),
	}
}

type exchangeCodeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

func (c *embeddedSignupClient) ExchangeCode(ctx context.Context, appID, appSecret, code string) (string, error) {
	if appID == "" || appSecret == "" {
		return "", fmt.Errorf("faltan app_id o app_secret en las credenciales de plataforma")
	}
	if code == "" {
		return "", fmt.Errorf("Meta no devolvió el código de autorización")
	}

	var result exchangeCodeResponse

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetQueryParam("client_id", appID).
		SetQueryParam("client_secret", appSecret).
		SetQueryParam("code", code).
		SetResult(&result).
		Get("oauth/access_token")
	if err != nil {
		return "", fmt.Errorf("error canjeando el código con Meta: %w", err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return "", parseMetaGraphError(resp.String(), resp.StatusCode(), 0)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("Meta no devolvió un token para la cuenta del negocio")
	}

	return result.AccessToken, nil
}

func (c *embeddedSignupClient) SubscribeApp(ctx context.Context, wabaID, accessToken string) error {
	if wabaID == "" {
		return fmt.Errorf("waba_id no configurado")
	}

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		Post(fmt.Sprintf("%s/subscribed_apps", wabaID))
	if err != nil {
		return fmt.Errorf("error suscribiendo la app al WABA %s: %w", wabaID, err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return parseMetaGraphError(resp.String(), resp.StatusCode(), 0)
	}

	return nil
}

func (c *embeddedSignupClient) ListPhoneNumbers(ctx context.Context, wabaID, accessToken string) ([]ports.WABAPhoneNumber, error) {
	return c.templates.ListPhoneNumbers(ctx, wabaID, accessToken)
}

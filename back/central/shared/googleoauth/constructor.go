package googleoauth

import (
	"context"
	"net/http"
	"time"

	"github.com/secamc93/probability/back/central/shared/env"
	"github.com/secamc93/probability/back/central/shared/log"
)

const (
	authEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenEndpoint = "https://oauth2.googleapis.com/token"
	scopes        = "openid email profile"
)

type Profile struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

type IClient interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*Profile, error)
	IsConfigured() bool
}

type client struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
	logger       log.ILogger
}

func New(cfg env.IConfig, logger log.ILogger) IClient {
	return &client{
		clientID:     cfg.Get("GOOGLE_OAUTH_CLIENT_ID"),
		clientSecret: cfg.Get("GOOGLE_OAUTH_CLIENT_SECRET"),
		redirectURI:  cfg.Get("GOOGLE_OAUTH_REDIRECT_URI"),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		logger:       logger,
	}
}

func (c *client) IsConfigured() bool {
	return c.clientID != "" && c.clientSecret != "" && c.redirectURI != ""
}

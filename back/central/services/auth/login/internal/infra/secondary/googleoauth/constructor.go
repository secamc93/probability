package googleoauth

import (
	"net/http"
	"time"

	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/shared/env"
	"github.com/secamc93/probability/back/central/shared/log"
)

const (
	authEndpoint  = "https://accounts.google.com/o/oauth2/v2/auth"
	tokenEndpoint = "https://oauth2.googleapis.com/token"
	scopes        = "openid email profile"
)

type Provider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
	logger       log.ILogger
}

func New(cfg env.IConfig, logger log.ILogger) domain.IGoogleOAuthProvider {
	return &Provider{
		clientID:     cfg.Get("GOOGLE_OAUTH_CLIENT_ID"),
		clientSecret: cfg.Get("GOOGLE_OAUTH_CLIENT_SECRET"),
		redirectURI:  cfg.Get("GOOGLE_OAUTH_REDIRECT_URI"),
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		logger:       logger,
	}
}

func (p *Provider) IsConfigured() bool {
	return p.clientID != "" && p.clientSecret != "" && p.redirectURI != ""
}

package usecaseembedded

import (
	"context"

	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/domain/ports"
	"github.com/secamc93/probability/back/central/shared/log"
)

const whatsAppTypeID uint = 2

type Config struct {
	Enabled      bool   `json:"enabled"`
	AppID        string `json:"app_id"`
	ConfigID     string `json:"config_id"`
	GraphVersion string `json:"graph_version"`
}

type SignupInput struct {
	Code          string
	WABAID        string
	PhoneNumberID string
}

type SignupResult struct {
	IntegrationID      uint
	BusinessID         uint
	WABAID             string
	PhoneNumberID      string
	DisplayPhoneNumber string
	VerifiedName       string
	QualityRating      string
	Registered         bool
	Pin                string
	Warning            string
}

type IUseCase interface {
	GetConfig(ctx context.Context) (*Config, error)
	Complete(ctx context.Context, businessID uint, input SignupInput) (*SignupResult, error)
}

type IUseCaseMutable interface {
	IUseCase
	SetResolver(resolver ports.IPlatformCredentialsGetter)
}

type SignupAPIFactory func(baseURL string) ports.IEmbeddedSignupAPI
type NumbersAPIFactory func(baseURL string) ports.IPhoneNumbersAPI

type usecase struct {
	credentialsCache ports.ICredentialsCache
	signupFactory    SignupAPIFactory
	numbersFactory   NumbersAPIFactory
	log              log.ILogger
	resolver         ports.IPlatformCredentialsGetter
}

func New(
	credentialsCache ports.ICredentialsCache,
	signupFactory SignupAPIFactory,
	numbersFactory NumbersAPIFactory,
	logger log.ILogger,
) IUseCaseMutable {
	return &usecase{
		credentialsCache: credentialsCache,
		signupFactory:    signupFactory,
		numbersFactory:   numbersFactory,
		log:              logger.WithModule("whatsapp-embedded-signup"),
	}
}

func (u *usecase) SetResolver(resolver ports.IPlatformCredentialsGetter) {
	u.resolver = resolver
}

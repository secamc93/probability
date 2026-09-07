package usecaseconnection

import (
	"context"

	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/domain/ports"
	"github.com/secamc93/probability/back/central/shared/log"
)

const whatsAppTypeID uint = 2

type SaveConnectionInput struct {
	UsePlatformToken bool
	WABAID           string
	PhoneNumberID    string
	AccessToken      string
}

type ConnectionResult struct {
	IntegrationID      uint
	BusinessID         uint
	OwnNumber          bool
	WABAID             string
	PhoneNumberID      string
	DisplayPhoneNumber string
	VerifiedName       string
	QualityRating      string
	PlatformToken      bool
	HostedByPlatform   bool
}

type IUseCase interface {
	SaveConnection(ctx context.Context, businessID uint, input SaveConnectionInput) (*ConnectionResult, error)
}

type IUseCaseMutable interface {
	IUseCase
	SetResolver(resolver ports.IPlatformCredentialsGetter)
}

type APIFactory func(baseURL string) ports.ITemplateAPI

type usecase struct {
	credentialsCache ports.ICredentialsCache
	apiFactory       APIFactory
	log              log.ILogger
	resolver         ports.IPlatformCredentialsGetter
}

func New(credentialsCache ports.ICredentialsCache, apiFactory APIFactory, logger log.ILogger) IUseCaseMutable {
	return &usecase{
		credentialsCache: credentialsCache,
		apiFactory:       apiFactory,
		log:              logger.WithModule("whatsapp-connection"),
	}
}

func (u *usecase) SetResolver(resolver ports.IPlatformCredentialsGetter) {
	u.resolver = resolver
}

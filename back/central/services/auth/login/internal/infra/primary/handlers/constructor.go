package authhandler

import (
	"github.com/secamc93/probability/back/central/services/auth/login/internal/app"
	"github.com/secamc93/probability/back/central/shared/log"

	"github.com/gin-gonic/gin"
)

type IAuthHandler interface {
	LoginHandler(c *gin.Context)
	VerifyHandler(c *gin.Context)
	GetUserRolesPermissionsHandler(c *gin.Context)
	ChangePasswordHandler(c *gin.Context)
	GeneratePasswordHandler(c *gin.Context)
	RecoveryChannelsHandler(c *gin.Context)
	ForgotPasswordHandler(c *gin.Context)
	VerifyOTPHandler(c *gin.Context)
	ResetPasswordHandler(c *gin.Context)
	GoogleAuthHandler(c *gin.Context)
	GoogleCallbackHandler(c *gin.Context)
	SessionHandler(c *gin.Context)
	RegisterRoutes(v1Group *gin.RouterGroup, handler IAuthHandler, logger log.ILogger)
}

type AuthHandler struct {
	usecase app.Iapp
	logger  log.ILogger
}

func New(usecase app.Iapp, logger log.ILogger) IAuthHandler {
	contextualLogger := logger.WithModule("autenticación")
	return &AuthHandler{
		usecase: usecase,
		logger:  contextualLogger,
	}
}

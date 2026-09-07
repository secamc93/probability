package login

import (
	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/app"
	authhandler "github.com/secamc93/probability/back/central/services/auth/login/internal/infra/primary/handlers"
	googleadapter "github.com/secamc93/probability/back/central/services/auth/login/internal/infra/secondary/google"
	otpqueue "github.com/secamc93/probability/back/central/services/auth/login/internal/infra/secondary/queue"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/infra/secondary/repository"
	"github.com/secamc93/probability/back/central/shared/db"
	"github.com/secamc93/probability/back/central/shared/email"
	"github.com/secamc93/probability/back/central/shared/env"
	"github.com/secamc93/probability/back/central/shared/googleoauth"
	"github.com/secamc93/probability/back/central/shared/jwt"
	"github.com/secamc93/probability/back/central/shared/log"
	"github.com/secamc93/probability/back/central/shared/rabbitmq"
)

func New(
	router *gin.RouterGroup,
	db db.IDatabase,
	logger log.ILogger,
	cfg env.IConfig,
	queue rabbitmq.IQueue,
) {
	repo := repository.New(db, logger)

	jwtService := jwt.New(cfg.Get("JWT_SECRET"))

	emailService := email.New(cfg, logger)

	otpPublisher := otpqueue.New(queue, logger)

	googleProvider := googleadapter.New(googleoauth.New(cfg, logger))

	authUC := app.New(repo, jwtService, emailService, otpPublisher, googleProvider, logger, cfg)

	authH := authhandler.New(authUC, logger)

	authH.RegisterRoutes(router, authH, logger)
}

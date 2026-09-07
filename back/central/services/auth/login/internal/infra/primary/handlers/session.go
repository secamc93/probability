package authhandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/infra/primary/handlers/mapper"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/infra/primary/handlers/response"
	"github.com/secamc93/probability/back/central/services/auth/middleware"
	"github.com/secamc93/probability/back/central/shared/log"
)

func (h *AuthHandler) SessionHandler(c *gin.Context) {
	ctx := log.WithFunctionCtx(c.Request.Context(), "SessionHandler")

	authInfo, exists := middleware.GetAuthInfo(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "No autorizado"})
		return
	}

	domainResponse, err := h.usecase.CurrentSession(ctx, authInfo.UserID)
	if err != nil {
		h.logger.Error(ctx).Err(err).Uint("user_id", authInfo.UserID).Msg("Error al reconstruir la sesion")
		if errors.Is(err, domain.ErrUserNotFound) || errors.Is(err, domain.ErrUserInactive) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Error interno del servidor"})
		return
	}

	sessionResponse := mapper.ToLoginResponse(domainResponse)
	sessionResponse.Token = ""

	c.JSON(http.StatusOK, response.LoginSuccessResponse{
		Success: true,
		Data:    *sessionResponse,
	})
}

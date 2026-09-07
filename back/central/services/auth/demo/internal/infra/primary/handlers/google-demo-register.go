package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/auth/demo/internal/domain"
	"github.com/secamc93/probability/back/central/services/auth/demo/internal/infra/primary/handlers/request"
	"github.com/secamc93/probability/back/central/shared/session"
)

func (h *Handler) DemoRegisterWithGoogleHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var req request.GoogleDemoRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error(ctx).Err(err).Msg("Datos invalidos en registro demo con Google")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos de entrada invalidos: " + err.Error()})
		return
	}

	resp, err := h.usecase.DemoRegisterWithGoogle(ctx, domain.GoogleDemoRegisterRequest{
		SignupToken:  req.SignupToken,
		BusinessName: req.BusinessName,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrGoogleSignupTokenInvalid):
			status = http.StatusUnauthorized
		case errors.Is(err, domain.ErrBusinessNameRequired):
			status = http.StatusBadRequest
		case errors.Is(err, domain.ErrEmailAlreadyRegistered):
			status = http.StatusConflict
		}
		h.logger.Error(ctx).Err(err).Msg("Error en registro demo con Google")
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	clientType := c.GetHeader("X-Client-Type")
	isMobileClient := clientType == "mobile" || clientType == "api"

	token := resp.Token
	if !isMobileClient {
		session.SetCookie(c, resp.Token)
		token = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
		"data": gin.H{
			"token":         token,
			"user_id":       resp.UserID,
			"business_id":   resp.BusinessID,
			"name":          resp.FullName,
			"email":         resp.Email,
			"business_name": resp.BusinessName,
			"avatar_url":    resp.AvatarURL,
		},
	})
}

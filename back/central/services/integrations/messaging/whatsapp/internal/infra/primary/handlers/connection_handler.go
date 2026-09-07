package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/app/usecaseconnection"
)

type saveConnectionRequest struct {
	UsePlatformToken bool   `json:"use_platform_token"`
	WABAID           string `json:"waba_id"`
	PhoneNumberID    string `json:"phone_number_id"`
	AccessToken      string `json:"access_token"`
}

func (h *handler) SaveConnection(c *gin.Context) {
	ctx := c.Request.Context()

	if h.connectionUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "el módulo de conexión de WhatsApp no está disponible",
		})
		return
	}

	businessID, ok := h.resolveBusinessID(c)
	if !ok {
		return
	}

	var req saveConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "payload inválido",
		})
		return
	}

	result, err := h.connectionUseCase.SaveConnection(ctx, businessID, usecaseconnection.SaveConnectionInput{
		UsePlatformToken: req.UsePlatformToken,
		WABAID:           req.WABAID,
		PhoneNumberID:    req.PhoneNumberID,
		AccessToken:      req.AccessToken,
	})
	if err != nil {
		h.log.Warn(ctx).Err(err).
			Uint("business_id", businessID).
			Msg("[WhatsApp Conexión] - no se pudo guardar la conexión")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"integration_id":       result.IntegrationID,
			"business_id":          result.BusinessID,
			"own_number":           result.OwnNumber,
			"waba_id":              result.WABAID,
			"phone_number_id":      result.PhoneNumberID,
			"display_phone_number": result.DisplayPhoneNumber,
			"verified_name":        result.VerifiedName,
			"quality_rating":       result.QualityRating,
			"platform_token":       result.PlatformToken,
		},
	})
}

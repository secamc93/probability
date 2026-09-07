package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/integrations/messaging/whatsapp/internal/app/usecaseembedded"
)

type embeddedSignupRequest struct {
	Code          string `json:"code"`
	WABAID        string `json:"waba_id"`
	PhoneNumberID string `json:"phone_number_id"`
}

func (h *handler) GetEmbeddedSignupConfig(c *gin.Context) {
	if h.embeddedUseCase == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"enabled": false}})
		return
	}

	config, err := h.embeddedUseCase.GetConfig(c.Request.Context())
	if err != nil {
		h.log.Warn(c.Request.Context()).Err(err).Msg("[WhatsApp Registro insertado] - no se pudo leer la configuración")
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"enabled": false}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"enabled":       config.Enabled,
			"app_id":        config.AppID,
			"config_id":     config.ConfigID,
			"graph_version": config.GraphVersion,
		},
	})
}

func (h *handler) CompleteEmbeddedSignup(c *gin.Context) {
	ctx := c.Request.Context()

	if h.embeddedUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "el registro insertado no está disponible",
		})
		return
	}

	businessID, ok := h.resolveBusinessID(c)
	if !ok {
		return
	}

	var req embeddedSignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "payload inválido"})
		return
	}

	result, err := h.embeddedUseCase.Complete(ctx, businessID, usecaseembedded.SignupInput{
		Code:          req.Code,
		WABAID:        req.WABAID,
		PhoneNumberID: req.PhoneNumberID,
	})
	if err != nil {
		h.log.Warn(ctx).Err(err).
			Uint("business_id", businessID).
			Msg("[WhatsApp Registro insertado] - no se pudo completar")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"integration_id":       result.IntegrationID,
			"business_id":          result.BusinessID,
			"waba_id":              result.WABAID,
			"phone_number_id":      result.PhoneNumberID,
			"display_phone_number": result.DisplayPhoneNumber,
			"verified_name":        result.VerifiedName,
			"quality_rating":       result.QualityRating,
			"registered":           result.Registered,
			"pin":                  result.Pin,
			"warning":              result.Warning,
		},
	})
}

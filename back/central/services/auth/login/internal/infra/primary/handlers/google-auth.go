package authhandler

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/shared/log"
)

const googleStateCookie = "g_oauth_state"

func (h *AuthHandler) GoogleAuthHandler(c *gin.Context) {
	ctx := log.WithFunctionCtx(c.Request.Context(), "GoogleAuthHandler")

	result, err := h.usecase.GoogleAuthURL(ctx)
	if err != nil {
		h.logger.Error(ctx).Err(err).Msg("No se pudo construir la URL de autorizacion de Google")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", err))
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(googleStateCookie, result.State, 600, "/", "", isSecureRequest(c), true)

	c.Redirect(http.StatusFound, result.AuthURL)
}

func (h *AuthHandler) GoogleCallbackHandler(c *gin.Context) {
	ctx := log.WithFunctionCtx(c.Request.Context(), "GoogleCallbackHandler")

	if googleErr := c.Query("error"); googleErr != "" {
		h.logger.Warn(ctx).Str("google_error", googleErr).Msg("Google devolvio un error en el callback")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", domain.ErrGoogleExchangeFailed))
		return
	}

	state := c.Query("state")
	stateCookie, err := c.Cookie(googleStateCookie)
	if err != nil || state == "" || state != stateCookie {
		h.logger.Warn(ctx).Msg("State de Google invalido o ausente")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", domain.ErrGoogleInvalidState))
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(googleStateCookie, "", -1, "/", "", isSecureRequest(c), true)

	domainResponse, err := h.usecase.LoginWithGoogle(ctx, domain.GoogleCallbackRequest{
		Code:  c.Query("code"),
		State: state,
	})
	if err != nil {
		h.logger.Error(ctx).Err(err).Msg("Error en el login con Google")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", googleUserFacingError(err)))
		return
	}

	setSessionCookie(c, domainResponse.Token)

	h.logger.Info(ctx).
		Uint("user_id", domainResponse.User.ID).
		Str("email", domainResponse.User.Email).
		Bool("is_super_admin", domainResponse.IsSuperAdmin).
		Msg("Login con Google exitoso")

	target := frontendBaseURL() + "/auth/google/callback?status=ok"
	if domainResponse.RequirePasswordChange {
		target += "&require_password_change=true"
	}
	c.Redirect(http.StatusFound, target)
}

func googleUserFacingError(err error) error {
	switch {
	case errors.Is(err, domain.ErrGoogleUserNotFound),
		errors.Is(err, domain.ErrGoogleEmailNotVerified),
		errors.Is(err, domain.ErrGoogleAccountLinkedElsewhere),
		errors.Is(err, domain.ErrGoogleNotConfigured),
		errors.Is(err, domain.ErrUserInactive),
		errors.Is(err, domain.ErrUserPendingVerification):
		return err
	default:
		return domain.ErrGoogleExchangeFailed
	}
}

func frontendBaseURL() string {
	base := strings.TrimRight(os.Getenv("FRONTEND_BASE_URL"), "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return base
}

func frontendRedirect(param string, err error) string {
	return frontendBaseURL() + "/login?" + param + "=" + url.QueryEscape(err.Error())
}

func isSecureRequest(c *gin.Context) bool {
	if c.Request.TLS != nil {
		return true
	}
	return c.GetHeader("X-Forwarded-Proto") == "https"
}

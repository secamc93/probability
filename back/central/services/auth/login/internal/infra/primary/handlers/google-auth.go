package authhandler

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/secamc93/probability/back/central/services/auth/login/internal/domain"
	"github.com/secamc93/probability/back/central/shared/log"
	"github.com/secamc93/probability/back/central/shared/session"
)

const (
	googleStateCookie  = "g_oauth_state"
	googleIntentCookie = "g_oauth_intent"
	googleSignupCookie = "g_signup_token"
	intentDemo         = "demo"
	signupTokenTTL     = 15 * time.Minute
)

func (h *AuthHandler) GoogleAuthHandler(c *gin.Context) {
	ctx := log.WithFunctionCtx(c.Request.Context(), "GoogleAuthHandler")

	result, err := h.usecase.GoogleAuthURL(ctx)
	if err != nil {
		h.logger.Error(ctx).Err(err).Msg("No se pudo construir la URL de autorizacion de Google")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", err))
		return
	}

	seguro := isSecureRequest(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(googleStateCookie, result.State, 600, "/", "", seguro, true)

	intent := ""
	if c.Query("intent") == intentDemo {
		intent = intentDemo
	}
	c.SetCookie(googleIntentCookie, intent, 600, "/", "", seguro, true)

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

	intent, _ := c.Cookie(googleIntentCookie)

	seguro := isSecureRequest(c)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(googleStateCookie, "", -1, "/", "", seguro, true)
	c.SetCookie(googleIntentCookie, "", -1, "/", "", seguro, true)

	result, err := h.usecase.LoginWithGoogle(ctx, domain.GoogleCallbackRequest{
		Code:  c.Query("code"),
		State: state,
	})
	if err != nil {
		h.logger.Error(ctx).Err(err).Msg("Error en el login con Google")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", googleUserFacingError(err)))
		return
	}

	if result.NeedsSignup {
		h.redirigirARegistroDemo(c, ctx, result.Profile, intent)
		return
	}

	session.SetCookie(c, result.Session.Token)

	h.logger.Info(ctx).
		Uint("user_id", result.Session.User.ID).
		Str("email", result.Session.User.Email).
		Bool("is_super_admin", result.Session.IsSuperAdmin).
		Msg("Login con Google exitoso")

	target := frontendBaseURL() + "/auth/google/callback?status=ok"
	if result.Session.RequirePasswordChange {
		target += "&require_password_change=true"
	}
	c.Redirect(http.StatusFound, target)
}

func (h *AuthHandler) redirigirARegistroDemo(c *gin.Context, ctx context.Context, profile *domain.GoogleProfile, intent string) {
	if intent != intentDemo {
		h.logger.Warn(ctx).Str("email", profile.Email).Msg("Login con Google sin cuenta y sin intencion de crear demo")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", cuentaInexistente(profile.Email)))
		return
	}

	token, err := h.usecase.GoogleSignupToken(ctx, profile, signupTokenTTL)
	if err != nil {
		h.logger.Error(ctx).Err(err).Msg("No se pudo generar el token de registro con Google")
		c.Redirect(http.StatusFound, frontendRedirect("google_error", domain.ErrGoogleExchangeFailed))
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(googleSignupCookie, token, int(signupTokenTTL.Seconds()), "/", "", isSecureRequest(c), true)

	h.logger.Info(ctx).Str("email", profile.Email).Msg("Cuenta de Google sin usuario: se ofrece crear demo")
	c.Redirect(http.StatusFound, frontendBaseURL()+"/registro-demo")
}

func cuentaInexistente(email string) error {
	return errors.New(domain.ErrGoogleUserNotFound.Error() + " " + email + ". Pide a un administrador que te cree el usuario")
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

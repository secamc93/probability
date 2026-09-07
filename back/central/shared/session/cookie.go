package session

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

const CookieName = "session_token"

const maxAgeSeconds = 7 * 24 * 60 * 60

func SetCookie(c *gin.Context, token string) {
	cookieDomain := os.Getenv("SESSION_COOKIE_DOMAIN")
	if cookieDomain == "" {
		cookieDomain = ".probabilityia.com.co"
	}

	var cookieValue string
	if cookieDomain == "none" {
		cookieValue = fmt.Sprintf(
			"%s=%s; Max-Age=%d; Path=%s; HttpOnly; SameSite=Lax",
			CookieName,
			token,
			maxAgeSeconds,
			"/",
		)
	} else {
		cookieValue = fmt.Sprintf(
			"%s=%s; Max-Age=%d; Path=%s; Domain=%s; Secure; HttpOnly; SameSite=None; Partitioned",
			CookieName,
			token,
			maxAgeSeconds,
			"/",
			cookieDomain,
		)
	}
	c.Header("Set-Cookie", cookieValue)
}

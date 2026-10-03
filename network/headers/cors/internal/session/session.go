// Package session holds demo login handlers shared by both servers
// (hand-rolled CORS and gin-contrib/cors): one login flow to compare.
package session

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Login sets an HttpOnly session cookie. JS can never read it;
// the browser attaches it automatically on later credentialed requests.
// Secure holds on localhost too (trustworthy origin); plain-HTTP LAN
// testing needs it off — keep this demo localhost-only instead.
func Login(c *gin.Context) {
	session, err := c.Cookie("session")
	if err != nil || session == "" {
		sesId, _ := uuid.NewV7()
		session = "demo-user-" + sesId.String()
		slog.Info("issuing new session", "new_id", session)
	} else {
		slog.Info("session reuse")
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    session,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
}

// Mode sets the session cookie with the requested SameSite level
// (strict|lax|none) for the cross-site demo (app.test → api.test).
// Secure defaults on (localhost is trustworthy); set COOKIE_INSECURE=1
// for plain-http non-localhost hosts (nip.io): Secure cookies die on
// arrival there, and the demo shows nothing.
func Mode(c *gin.Context) {
	mode := c.Param("mode")

	var sameSite http.SameSite

	switch mode {
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		mode = "lax"
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    "demo-user-1",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   os.Getenv("COOKIE_INSECURE") == "",
		SameSite: sameSite,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged in", "samesite": mode})
}

// Me proves the server received the cookie: no cookie, 401.
func Me(c *gin.Context) {
	session, err := c.Cookie("session")
	if err != nil || session == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no session"})

		return
	}

	c.JSON(http.StatusOK, gin.H{"user": session})
}

// Whoami renders plain text for top-level link clicks: fishermen see
// the user, strangers see 401. No JSON — viewable in the address bar.
func Whoami(c *gin.Context) {
	session, err := c.Cookie("session")
	if err != nil || session == "" {
		c.String(http.StatusUnauthorized, "no session (cookie did not travel)")

		return
	}

	c.String(http.StatusOK, "user: "+session)
}

// Package session holds demo login handlers shared by both servers
// (hand-rolled CORS and gin-contrib/cors): one login flow to compare.
package session

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Login sets an HttpOnly session cookie. JS can never read it;
// the browser attaches it automatically on later credentialed requests.
// Secure holds on localhost too (trustworthy origin); plain-HTTP LAN
// testing needs it off — keep this demo localhost-only instead.
func Login(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session",
		Value:    "demo-user-1",
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	c.JSON(http.StatusOK, gin.H{"message": "logged in"})
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

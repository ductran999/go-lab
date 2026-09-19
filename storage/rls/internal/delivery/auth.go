package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// tenantIDKey is the gin context key carrying the authenticated tenant.
const tenantIDKey = "tenantID"

var (
	// ErrMissingToken is returned when no bearer token is provided.
	ErrMissingToken = errors.New("missing bearer token")
	// ErrMalformedToken is returned when the token shape or claims are invalid.
	ErrMalformedToken = errors.New("malformed token")
	// ErrInvalidSignature is returned when the HMAC does not verify.
	ErrInvalidSignature = errors.New("invalid token signature")
)

// tenantClaims is the subset of JWT claims this server cares about.
type tenantClaims struct {
	TenantID int `json:"tenant_id"`
}

// tenantFromJWT verifies an HS256 Authorization header with secret and
// returns the tenant_id claim. Standard library only, no new dependency.
func tenantFromJWT(header, secret string) (int, error) {
	token, found := strings.CutPrefix(header, "Bearer ")
	if !found || token == "" {
		return 0, ErrMissingToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, ErrMalformedToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, ErrMalformedToken
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return 0, ErrMalformedToken
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))

	if !hmac.Equal(mac.Sum(nil), sig) {
		return 0, ErrInvalidSignature
	}

	var claims tenantClaims

	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return 0, ErrMalformedToken
	}

	if claims.TenantID <= 0 {
		return 0, ErrMalformedToken
	}

	return claims.TenantID, nil
}

// RequireTenant is gin middleware authenticating the tenant from JWT
// and injecting it into the request context. No tenant travels in
// URL params or query strings on protected routes.
func RequireTenant(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := tenantFromJWT(c.GetHeader("Authorization"), secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token"})

			return
		}

		c.Set(tenantIDKey, tenantID)
		c.Next()
	}
}

// tenantFromContext reads the tenant injected by RequireTenant.
func tenantFromContext(c *gin.Context) (int, bool) {
	v, ok := c.Get(tenantIDKey)
	if !ok {
		return 0, false
	}

	id, ok := v.(int)
	if !ok || id <= 0 {
		return 0, false
	}

	return id, true
}

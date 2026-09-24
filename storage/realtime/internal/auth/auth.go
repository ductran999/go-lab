// Package auth verifies HS256 bearer tokens carrying a tenant_id
// claim. Standard library only. EventSource sets no headers, so the
// token arrives via Authorization header (writers) or ?token=
// (browser streams) — both verified identically.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ductran999/shared-pkg/environ"
)

var (
	ErrMissingToken     = errors.New("missing token")
	ErrMalformedToken   = errors.New("malformed token")
	ErrInvalidSignature = errors.New("invalid token signature")
)

type tenantClaims struct {
	TenantID int `json:"tenant_id"`
}

func secret() string {
	return environ.Get("REALTIME_JWT_SECRET", "demo-secret-change-me-32-chars-min")
}

func verify(token, sec string) (int, error) {
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

	mac := hmac.New(sha256.New, []byte(sec))
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

// Mint issues a demo token for a tenant. No expiry, no rotation:
// demo-grade issuance for a lab, not a pattern to copy.
func Mint(tenantID int) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload, _ := json.Marshal(tenantClaims{TenantID: tenantID})
	body := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, []byte(secret()))
	_, _ = mac.Write([]byte(header + "." + body))

	return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// TokenFromRequest reads Bearer <t>, falling back to ?token= for
// EventSource streams that cannot set headers.
func TokenFromRequest(c *gin.Context) string {
	if token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer "); found && token != "" {
		return token
	}

	return c.Query("token")
}

// TenantID verifies the request token and returns its tenant.
func TenantID(c *gin.Context) (int, error) {
	token := TokenFromRequest(c)
	if token == "" {
		return 0, ErrMissingToken
	}

	return verify(token, secret())
}

// RequireTenant aborts 401 unless a valid token names a tenant.
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, err := TenantID(c)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or missing token"})

			return
		}

		c.Set("tenantID", tenantID)
		c.Next()
	}
}

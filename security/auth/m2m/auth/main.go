package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey = []byte("m2m_shared_secret_key_123")

// APIClient represents a registered system/application, not a human user
type APIClient struct {
	ClientID      string
	ClientSecret  string
	AllowedScopes []string // Permitted scopes for this specific system
}

// Mock database of registered API Clients
var registeredClients = map[string]APIClient{
	"billing_service": {
		ClientID:      "billing_service",
		ClientSecret:  "billing_secret_999",
		AllowedScopes: []string{"read:reports", "write:billing"},
	},
}

func main() {
	r := gin.Default()

	// ---------------------------------------------------------
	// 1. ENDPOINT: POST /oauth/token (Client Credentials Grant)
	// ---------------------------------------------------------
	r.POST("/oauth/token", func(c *gin.Context) {
		// OAuth 2.0 spec mandates checking the grant_type
		grantType := c.PostForm("grant_type")
		if grantType != "client_credentials" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
			return
		}

		clientID := c.PostForm("client_id")
		clientSecret := c.PostForm("client_secret")

		// Authenticate the calling system
		client, exists := registeredClients[clientID]
		if !exists || client.ClientSecret != clientSecret {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_client"})
			return
		}

		// Generate M2M JWT (Notice: No User ID is involved here)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":    "auth_server",
			"sub":    client.ClientID, // Subject is the client application itself
			"scopes": client.AllowedScopes,
			"exp":    time.Now().Add(time.Hour * 1).Unix(), // Valid for 1 hour
		})
		tokenString, _ := token.SignedString(jwtSecretKey)

		// Return standard OAuth 2.0 token response
		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   3600,
			"scope":        strings.Join(client.AllowedScopes, " "),
		})
	})

	// ---------------------------------------------------------
	// 2. RESOURCE API: GET /api/v1/reports
	// ---------------------------------------------------------
	r.GET("/api/v1/reports", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify the Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		// Extract Scopes from JWT Claims
		claims, _ := token.Claims.(jwt.MapClaims)
		scopes, _ := claims["scopes"].([]any)

		// --- SCOPE AUTHORIZATION CHECK ---
		// Verify if this calling system has the required "read:reports" scope
		hasRequiredScope := false
		for _, s := range scopes {
			if s.(string) == "read:reports" {
				hasRequiredScope = true
				break
			}
		}

		if !hasRequiredScope {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: Missing 'read:reports' scope"})
			return
		}

		// Return the protected business data
		c.JSON(http.StatusOK, gin.H{
			"generated_at":   time.Now().Format(time.RFC3339),
			"revenue_report": "$1,250,000",
			"requested_by":   claims["sub"], // Confirms which system called us
		})
	})

	r.Run(":8080")
}

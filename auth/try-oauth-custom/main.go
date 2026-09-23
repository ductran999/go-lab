package main

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

//go:embed templates/*
var templateFS embed.FS

var jwtSecretKey = []byte("super_secret_jwt_key_123")

type RegisteredClient struct {
	ClientID    string
	AllowedURIs []string
}

var registeredClients = map[string]RegisteredClient{
	"my_spa_client": {
		ClientID:    "my_spa_client",
		AllowedURIs: []string{"http://localhost:8080/oauth/callback"},
	},
}

// AuthSession stores the temporary authorization code details
type AuthSession struct {
	UserID        string
	ClientID      string
	RedirectURI   string
	CodeChallenge string // Store the PKCE Lock
	ExpiresAt     time.Time
}

var authCodeStore = make(map[string]AuthSession)

// Verify PKCE: SHA256(verifier) == challenge
func verifyPKCE(verifier, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	return encoded == challenge
}

func main() {
	r := gin.Default()

	// Load templates from Go Embedded Filesystem
	templ := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(templ)

	// 1. Home Page (Serve the Frontend Client)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"BackendURL": "http://localhost:8080",
		})
	})

	// 2. OAuth2 Callback Helper Page
	r.GET("/oauth/callback", func(c *gin.Context) {
		c.HTML(http.StatusOK, "callback.html", nil)
	})

	// 3. OAuth2 Authorize Consent Screen (Popup opens here)
	r.GET("/oauth/authorize", func(c *gin.Context) {
		clientID := c.Query("client_id")
		redirectURI := c.Query("redirect_uri")
		codeChallenge := c.Query("code_challenge")
		state := c.Query("state")

		// Check client_id is registered
		client, exists := registeredClients[clientID]
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_client: Client ID is not registered"})
			return
		}

		if !slices.Contains(client.AllowedURIs, redirectURI) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_redirect_uri: Redirect URI mismatch"})
			return
		}

		c.HTML(http.StatusOK, "authorize.html", gin.H{
			"ClientID":      clientID,
			"RedirectURI":   redirectURI,
			"CodeChallenge": codeChallenge,
			"State":         state,
		})
	})

	// 4. OAuth2 Approve Action (User clicked "Approve" in Popup)
	r.GET("/oauth/authorize/approve", func(c *gin.Context) {
		clientID := c.Query("client_id")
		redirectURI := c.Query("redirect_uri")
		codeChallenge := c.Query("code_challenge")
		state := c.Query("state")

		// Mock authenticated user ID from internal DB
		mockUserID := "user_999888"

		// Generate one-time authorization code
		authCode := uuid.New().String()

		// Store in-memory for 5 minutes
		authCodeStore[authCode] = AuthSession{
			UserID:        mockUserID,
			ClientID:      clientID,
			RedirectURI:   redirectURI,
			CodeChallenge: codeChallenge,
			ExpiresAt:     time.Now().Add(5 * time.Minute),
		}

		// Redirect back to client callback with code
		callbackURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state)
		c.Redirect(http.StatusFound, callbackURL)
	})

	// 5. OAuth2 Token Exchange with PKCE Verification
	r.POST("/oauth/token", func(c *gin.Context) {
		code := c.PostForm("code")
		clientID := c.PostForm("client_id")
		codeVerifier := c.PostForm("code_verifier") // Plain key from client

		session, exists := authCodeStore[code]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization code"})
			return
		}

		if time.Now().After(session.ExpiresAt) {
			delete(authCodeStore, code)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization code expired"})
			return
		}

		if session.ClientID != clientID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Client ID mismatch"})
			return
		}

		// ---> PKCE VERIFICATION STEP <---
		if !verifyPKCE(codeVerifier, session.CodeChallenge) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "PKCE verification failed: Invalid code_verifier"})
			return
		}

		// Valid! Clean up the used code (Prevent reuse)
		delete(authCodeStore, code)

		// Generate custom system JWT for the authenticated user
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": "oauth2_dev@example.com",
			"name":  "OAuth2 Go Developer",
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, _ := token.SignedString(jwtSecretKey)

		// Return the token payload to the client
		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	// 6. Resource Server API (Profile Endpoint)
	r.GET("/api/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid token"})
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		c.JSON(http.StatusOK, gin.H{
			"name":  claims["name"],
			"email": claims["email"],
		})
	})

	r.Run(":8080")
}

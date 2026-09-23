package main

import (
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

//go:embed templates/*
var templateFS embed.FS

var jwtSecretKey = []byte("super_secret_jwt_key_123")

type AuthSession struct {
	UserID        string
	ClientID      string
	RedirectURI   string
	CodeChallenge string
	ExpiresAt     time.Time
}

var authCodeStore = make(map[string]AuthSession)

func verifyPKCE(verifier, challenge string) bool {
	hash := sha256.Sum256([]byte(verifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	return encoded == challenge
}

func main() {
	r := gin.Default()

	templ := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(templ)

	// 1. Home Page (Serve Frontend Client)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"BackendURL": "http://localhost:8080",
		})
	})

	// 2. OAuth2 Callback Helper Page
	r.GET("/oauth/callback", func(c *gin.Context) {
		c.HTML(http.StatusOK, "callback.html", nil)
	})

	// 3. OIDC /oauth/authorize Page (Consent Screen)
	r.GET("/oauth/authorize", func(c *gin.Context) {
		c.HTML(http.StatusOK, "authorize.html", gin.H{
			"ClientID":      c.Query("client_id"),
			"RedirectURI":   c.Query("redirect_uri"),
			"CodeChallenge": c.Query("code_challenge"),
			"State":         c.Query("state"),
		})
	})

	// 4. OIDC Approve Action
	r.GET("/oauth/authorize/approve", func(c *gin.Context) {
		clientID := c.Query("client_id")
		redirectURI := c.Query("redirect_uri")
		codeChallenge := c.Query("code_challenge")
		state := c.Query("state")

		mockUserID := "user_999888"

		authCode := uuid.New().String()

		authCodeStore[authCode] = AuthSession{
			UserID:        mockUserID,
			ClientID:      clientID,
			RedirectURI:   redirectURI,
			CodeChallenge: codeChallenge,
			ExpiresAt:     time.Now().Add(5 * time.Minute),
		}

		callbackURL := fmt.Sprintf("%s?code=%s&state=%s", redirectURI, authCode, state)
		c.Redirect(http.StatusFound, callbackURL)
	})

	// 5. OIDC Token Exchange (Returns Access Token AND ID Token)
	r.POST("/oauth/token", func(c *gin.Context) {
		code := c.PostForm("code")
		clientID := c.PostForm("client_id")
		codeVerifier := c.PostForm("code_verifier")

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

		// PKCE Verification
		if !verifyPKCE(codeVerifier, session.CodeChallenge) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "PKCE verification failed"})
			return
		}

		delete(authCodeStore, code)

		// A. Generate Access Token (Dành cho máy móc / APIs)
		accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":   session.UserID,
			"scope": "openid profile email",
			"exp":   time.Now().Add(time.Hour * 1).Unix(),
		})
		accessTokenString, _ := accessToken.SignedString(jwtSecretKey)

		// B. Generate OIDC ID Token (Identity Token - Dành cho con người / Giao diện)
		idToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":   "http://localhost:8080", // Ai phát hành
			"sub":   session.UserID,          // ID người dùng (Subject)
			"aud":   session.ClientID,        // Ai được dùng (Audience)
			"name":  "OIDC Master Developer", // Giao diện bóc trực tiếp cái này ra vẽ
			"email": "oidc_pro@example.com",
			"exp":   time.Now().Add(time.Hour * 1).Unix(),
		})
		idTokenString, _ := idToken.SignedString(jwtSecretKey)

		// Trả về bộ đôi token chuẩn chỉ OIDC
		c.JSON(http.StatusOK, gin.H{
			"access_token": accessTokenString,
			"id_token":     idTokenString, // Trả thêm ID Token chuẩn OIDC!
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})

	// 6. Standard OIDC Discovery Endpoint
	r.GET("/.well-known/openid-configuration", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"issuer":                   "http://localhost:8080",
			"authorization_endpoint":   "http://localhost:8080/oauth/authorize",
			"token_endpoint":           "http://localhost:8080/oauth/token",
			"userinfo_endpoint":        "http://localhost:8080/oauth/userinfo",
			"scopes_supported":         []string{"openid", "profile", "email"},
			"response_types_supported": []string{"code"},
			"subject_types_supported":  []string{"public"},
		})
	})

	r.Run(":8080")
}

package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base64"
	"html/template"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

//go:embed templates/*
var templateFS embed.FS

const (
	keycloakURL = "http://localhost:8080/realms/go-app-example"
	redirectURL = "http://localhost:8081/oauth/callback"
)

func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func main() {
	// Playground: export GO_BACKEND_CLIENT_ID/SECRET (Keycloak realm)
	// before running.
	clientID := os.Getenv("GO_BACKEND_CLIENT_ID")
	clientSecret := os.Getenv("GO_BACKEND_CLIENT_SECRET")

	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, keycloakURL)
	if err != nil {
		panic("Lỗi không thể kết nối đến Keycloak: " + err.Error())
	}

	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	r := gin.Default()
	templ := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(templ)

	// 1. Home Page (Serve Frontend Client)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"BackendURL": "http://localhost:8081",
		})
	})

	r.GET("/login", func(c *gin.Context) {
		state := generateRandomState()
		c.SetCookie("oauth_state", state, 300, "/", "localhost", false, true)

		url := oauth2Config.AuthCodeURL(state)
		c.Redirect(http.StatusFound, url)
	})

	r.GET("/oauth/callback", func(c *gin.Context) {
		cookieState, err := c.Cookie("oauth_state")
		if err != nil || c.Query("state") != cookieState {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Lỗi State: Phát hiện dấu hiệu tấn công CSRF!"})
			return
		}

		// B. Lấy mã code đổi lấy bộ Token
		oauth2Token, err := oauth2Config.Exchange(ctx, c.Query("code"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi khi đổi code lấy token: " + err.Error()})
			return
		}

		// C. Ép kiểu và lấy ID Token ra khỏi gói kết quả
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Không tìm thấy ID Token chuẩn OIDC!"})
			return
		}

		// D. XÁC THỰC ID TOKEN BẰNG PUBLIC KEY CỦA KEYCLOAK (Tự động hoàn toàn)
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Chữ ký ID Token không hợp lệ: " + err.Error()})
			return
		}

		// E. Bóc tách thông tin người dùng (Claims)
		var claims struct {
			Email             string `json:"email"`
			Verified          bool   `json:"email_verified"`
			Name              string `json:"name"`
			PreferredUsername string `json:"preferred_username"`
		}
		if err := idToken.Claims(&claims); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi đọc thông tin User: " + err.Error()})
			return
		}

		c.HTML(http.StatusOK, "callback.html", gin.H{
			"Message":    "Login Success fully!",
			"User":       claims,
			"IDToken":    rawIDToken,
			"BackendURL": "http://localhost:8081",
		})
	})

	r.Run(":8081")
}

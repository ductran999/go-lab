package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

//go:embed templates/*
var templateFS embed.FS

func main() {
	// Playground: export JWT_SESSION_KEY and JWT_SECRET_KEY (or drop a
	// .env next to a loader of your choice) before running.
	jwtSessionKey := []byte(os.Getenv("JWT_SESSION_KEY"))
	jwtSecretKey := []byte(os.Getenv("JWT_SECRET_KEY"))

	store := sessions.NewCookieStore(jwtSessionKey)
	store.MaxAge(86400 * 30)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	gothic.Store = store

	goth.UseProviders(
		google.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			os.Getenv("GOOGLE_REDIRECT_URL"),
			"email", "profile",
		),
	)

	r := gin.Default()
	templ := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"BackendURL": os.Getenv("BACKEND_URL"),
		})
	})

	r.GET("/auth/:provider", func(c *gin.Context) {
		provider := c.Param("provider")
		q := c.Request.URL.Query()
		q.Add("provider", provider)
		c.Request.URL.RawQuery = q.Encode()
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	r.GET("/auth/:provider/callback", func(c *gin.Context) {
		provider := c.Param("provider")
		q := c.Request.URL.Query()
		q.Add("provider", provider)
		c.Request.URL.RawQuery = q.Encode()

		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Got Error: " + err.Error()})
			return
		}

		// Create a JWT token with user information
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"email": user.Email,
			"name":  user.Name,
			"exp":   time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, _ := token.SignedString(jwtSecretKey)

		c.HTML(http.StatusOK, "callback.html", gin.H{
			"Token":      tokenString,
			"BackendURL": os.Getenv("BACKEND_URL"),
		})
	})

	r.GET("/api/profile", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No token provided"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecretKey, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid claims"})
			return
		}

		name, _ := claims["name"].(string)
		email, _ := claims["email"].(string)

		c.JSON(http.StatusOK, gin.H{
			"name":  name,
			"email": email,
		})
	})

	log.Println("Server is listening at", os.Getenv("SERVER_HOST"))
	r.Run(os.Getenv("SERVER_HOST"))
}

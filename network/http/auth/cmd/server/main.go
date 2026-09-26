// Command server demonstrates where tokens travel: Authorization
// header (strict), HttpOnly cookie, or query param. /me reports
// which channel carried the token; /me-strict accepts Bearer only.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ductran999/shared-pkg/environ"
)

func fail(err error) {
	slog.Error("server failed", "error", err)

	os.Exit(1)
}

var (
	ErrMissingToken     = errors.New("missing token")
	ErrMalformedToken   = errors.New("malformed token")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrExpiredToken     = errors.New("expired token")
)

type claims struct {
	User string `json:"user"`
	Exp  int64  `json:"exp"`
}

func secret() string {
	return environ.Get("AUTH_SECRET", "demo-secret-change-me-32-chars-min")
}

// sign mints an HS256 token valid for one hour. Demo-grade: no key
// rotation, no refresh — the transport lesson matters, not the PKI.
func sign(user string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payload, _ := json.Marshal(claims{User: user, Exp: time.Now().Add(time.Hour).Unix()})
	body := base64.RawURLEncoding.EncodeToString(payload)

	mac := hmac.New(sha256.New, []byte(secret()))
	_, _ = mac.Write([]byte(header + "." + body))

	return header + "." + body + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ErrMalformedToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", ErrMalformedToken
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return "", ErrMalformedToken
	}

	mac := hmac.New(sha256.New, []byte(secret()))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))

	if !hmac.Equal(mac.Sum(nil), sig) {
		return "", ErrInvalidSignature
	}

	var c claims

	err = json.Unmarshal(payload, &c)
	if err != nil {
		return "", ErrMalformedToken
	}

	if c.User == "" || time.Now().Unix() > c.Exp {
		return "", ErrExpiredToken
	}

	return c.User, nil
}

// bearer extracts the token from Authorization: Bearer <t>.
func bearer(r *http.Request) string {
	token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")

	if !found || token == "" {
		return ""
	}

	return token
}

// fromAny accepts header, cookie, then query — reporting the channel.
func fromAny(r *http.Request) (string, string) {
	if t := bearer(r); t != "" {
		return t, "header"
	}

	c, cookieErr := r.Cookie("token")
	if cookieErr == nil && c.Value != "" {
		return c.Value, "cookie"
	}

	if t := r.URL.Query().Get("token"); t != "" {
		return t, "query"
	}

	return "", ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8097")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)

			return
		}

		next(w, r)
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		User string `json:"user"`
	}

	_ = json.NewDecoder(r.Body).Decode(&body)

	if body.User == "" {
		body.User = "demo"
	}

	token := sign(body.User)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		// Cross-port counts as cross-site: None + Secure (localhost
		// counts as trustworthy). May still be rejected under
		// third-party-cookie blocking — that rejection IS the lesson.
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func me(w http.ResponseWriter, r *http.Request) {
	token, via := fromAny(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": ErrMissingToken.Error()})

		return
	}

	user, err := verify(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"user": user, "via": via})
}

func meStrict(w http.ResponseWriter, r *http.Request) {
	token := bearer(r)
	if token == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": ErrMissingToken.Error()})

		return
	}

	user, err := verify(token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})

		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"user": user, "via": "header"})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", cors(login))
	mux.HandleFunc("/me", cors(me))
	mux.HandleFunc("/me-strict", cors(meStrict))

	addr := ":" + environ.Get("PORT", "8098")

	slog.Info("serving auth lab", "addr", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	err := server.ListenAndServe()
	if err != nil {
		fail(err)
	}
}

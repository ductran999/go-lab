// Package middleware holds shared HTTP middleware: logging and demo auth.
// Chain as logging(auth(mux)) — stdlib has no Use(), wrapping is it.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel/trace"

	"go-lab/observability/otel/internal/requestid"
	"go-lab/observability/otel/internal/tracing"
)

// ctxKeys are unexported: no package can collide or forge our values.
type ctxKey string

const (
	tenantKey ctxKey = "tenantID"
	userKey   ctxKey = "userID"
)

// Logging is the single per-request log line: method/path/status/duration
// plus live trace.id + span.id plus tenant/user stamped by Auth plus the
// incoming parent.trace.id. Place it INSIDE otelhttp (span exists) and
// INSIDE Auth (identity exists): otelhttp → Auth → Logging → mux.
// Unauthenticated requests log nothing by design.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &recorder{writer: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		sc := trace.SpanFromContext(r.Context()).SpanContext()

		slog.Info("http",
			"method", r.Method, "path", r.URL.Path,
			"status", rec.status, "dur", time.Since(start).Round(time.Millisecond),
			"trace.id", sc.TraceID().String(), "span.id", sc.SpanID().String(),
			"tenant.id", TenantOf(r.Context()), "user.id", UserOf(r.Context()),
			"parent.trace.id", tracing.TraceIDOf(r, r.Context()),
			"req.id", requestid.Of(r.Context()),
		)
	})
}

// Auth verifies an HS256 JWT (tenant_id + user_id claims), then stamps
// both into the context and the current span (delivery owns observability).
// Demo secret from env; Mint issues lab tokens (see README).
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		claims, err := verifyJWT(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		ctx := context.WithValue(r.Context(), tenantKey, claims.TenantID)
		ctx = context.WithValue(ctx, userKey, claims.UserID)

		tracing.Business(ctx, claims.TenantID, claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// claims carries the identity this demo trusts from JWT.
//
//nolint:embeddedstructfieldcheck // jwt.RegisteredClaims embedding is the library's idiom.
type claims struct {
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
	jwt.RegisteredClaims
}

func jwtSecret() []byte {
	secret := os.Getenv("OTEL_JWT_SECRET")
	if secret == "" {
		secret = "demo-secret-change-me-32-chars-min"
	}

	return []byte(secret)
}

// verifyJWT parses and validates an HS256 token (expiry enforced).
func verifyJWT(token string) (*claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}

		return jwtSecret(), nil
	})
	if err != nil {
		return nil, err
	}

	c, ok := parsed.Claims.(*claims)
	if !ok || c.TenantID == "" || c.UserID == "" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return c, nil
}

// Mint issues a 1-hour JWT for demos (see README for the curl).
func Mint(tenantID, userID string) (string, error) {
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		TenantID: tenantID,
		UserID:   userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	return token.SignedString(jwtSecret())
}

// TenantOf reads the tenant stamped by Auth (empty when absent).
func TenantOf(ctx context.Context) string {
	tenant, _ := ctx.Value(tenantKey).(string)

	return tenant
}

// UserOf reads the user stamped by Auth (empty when absent).
func UserOf(ctx context.Context) string {
	user, _ := ctx.Value(userKey).(string)

	return user
}

type recorder struct {
	writer http.ResponseWriter
	status int
}

func (r *recorder) Header() http.Header {
	return r.writer.Header()
}

func (r *recorder) Write(b []byte) (int, error) {
	return r.writer.Write(b)
}

func (r *recorder) WriteHeader(status int) {
	r.status = status
	r.writer.WriteHeader(status)
}

// Flush forwards so streamed responses (SSE) keep working behind logging.
func (r *recorder) Flush() {
	if f, ok := r.writer.(http.Flusher); ok {
		f.Flush()
	}
}

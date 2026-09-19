package delivery

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"go-lab/storage/rls/internal/domain"
)

const testSecret = "test-secret-at-least-32-chars-long!!"

// signToken crafts an HS256 token for tests (same construction as make token).
func signToken(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	payloadBytes, err := json.Marshal(claims)
	require.NoError(t, err)

	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return header + "." + payload + "." + sig
}

func TestTenantFromJWTTable(t *testing.T) {
	valid := signToken(t, testSecret, map[string]any{"tenant_id": 1})
	wrongKey := signToken(t, "wrong-secret-00000000000000000000", map[string]any{"tenant_id": 1})
	noClaim := signToken(t, testSecret, map[string]any{"role": "x"})
	zeroTenant := signToken(t, testSecret, map[string]any{"tenant_id": 0})

	cases := []struct {
		name    string
		header  string
		wantID  int
		wantErr error
	}{
		{"valid", "Bearer " + valid, 1, nil},
		{"missing", "", 0, ErrMissingToken},
		{"no bearer prefix", valid, 0, ErrMissingToken},
		{"bad shape", "Bearer a.b", 0, ErrMalformedToken},
		{"bad signature", "Bearer " + wrongKey, 0, ErrInvalidSignature},
		{"missing claim", "Bearer " + noClaim, 0, ErrMalformedToken},
		{"zero tenant", "Bearer " + zeroTenant, 0, ErrMalformedToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := tenantFromJWT(tc.header, testSecret)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.wantID, id)
		})
	}
}

func serveRLS(t *testing.T, repo domain.DocumentRepository, secret, target, token string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)

	r := gin.New()
	NewRLSDocumentHandler(repo, secret).RegisterRLSRoutes(r.Group("/rls"))

	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestRLSRouteTenantFromToken(t *testing.T) {
	repo := stubRepository{byTenant: map[int][]domain.Document{
		1: {{ID: 1, TenantID: 1, Body: "a"}},
		2: {{ID: 2, TenantID: 2, Body: "b"}},
	}}

	token := signToken(t, testSecret, map[string]any{"tenant_id": 2})

	w := serveRLS(t, repo, testSecret, "/rls/docs", token)
	require.Equal(t, http.StatusOK, w.Code)

	var body []DocumentDTO
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, 2, body[0].TenantID)
}

func TestRLSRouteRejectsBadTokens(t *testing.T) {
	repo := stubRepository{}

	token := signToken(t, testSecret, map[string]any{"tenant_id": 1})

	w := serveRLS(t, repo, "other-secret-00000000000000000000", "/rls/docs", token)
	require.Equal(t, http.StatusUnauthorized, w.Code)

	w = serveRLS(t, repo, testSecret, "/rls/docs", "")
	require.Equal(t, http.StatusUnauthorized, w.Code)
}

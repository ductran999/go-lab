// Lesson 6: testing. httptest drives handlers in-process: no ports,
// no servers, table-driven cases. Run: go test ./lessons/06-testing/ -v.
package lesson06

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func todosHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	_ = json.NewEncoder(w).Encode([]string{"learn mux"})
}

func TestTodos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantType   string
	}{
		{name: "get ok", method: http.MethodGet, wantStatus: http.StatusOK, wantType: "application/json; charset=utf-8"},
		{name: "post rejected", method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(tt.method, "/todos", nil)
			rec := httptest.NewRecorder()

			todosHandler(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantType != "" && rec.Header().Get("Content-Type") != tt.wantType {
				t.Fatalf("content-type = %q, want %q", rec.Header().Get("Content-Type"), tt.wantType)
			}
		})
	}
}

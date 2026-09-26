// Package requestid ensures every edge request carries an X-Request-Id:
// incoming value is honored, otherwise a uuid is minted. It travels
// downstream as a header and lands in logs — the join key that works
// even where traces don't reach.
package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ctxKey is unexported: no package can collide or forge our values.
type ctxKey string

const idKey ctxKey = "requestID"

// Ensure stamps the request: reuse incoming id or mint one, mirror it
// back in the response so callers can search by it.
func Ensure(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}

		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), idKey, id)))
	})
}

// Of reads the id stamped by Ensure (empty when absent).
func Of(ctx context.Context) string {
	id, _ := ctx.Value(idKey).(string)

	return id
}

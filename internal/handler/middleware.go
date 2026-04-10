package handler

import (
	"context"
	"net/http"
	"strings"

	"wishlist/internal/auth"
)

type ctxKey int

const userIDKey ctxKey = 1

func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(userIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

func AuthMiddleware(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				Error(w, http.StatusUnauthorized, "missing or invalid authorization")
				return
			}
			raw := strings.TrimSpace(h[7:])
			uid, err := auth.ParseToken(raw, secret)
			if err != nil {
				Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

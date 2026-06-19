package middleware

import (
	"context"
	"net/http"
	"strings"

	"log/slog"
)

const authUserKey contextKey = 1

// TokenParser validates a bearer token and returns the subject (user id). It is
// provided by the caller (the auth package) to keep this package decoupled from
// JWT libraries.
type TokenParser func(token string) (userID string, err error)

// AuthRequired rejects requests without a valid Bearer token, otherwise stores
// the authenticated user id in the request context.
func AuthRequired(parser TokenParser, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
				return
			}
			userID, err := parser(token)
			if err != nil {
				log.Debug("auth: token rejected", "error", err)
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), authUserKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFrom returns the authenticated user id from the context, if any.
func UserIDFrom(ctx context.Context) string {
	if v, ok := ctx.Value(authUserKey).(string); ok {
		return v
	}
	return ""
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

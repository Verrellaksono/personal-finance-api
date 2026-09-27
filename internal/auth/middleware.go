package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

type MiddlewareContract interface {
	ValidateToken(tokenString string) (*UserClaims, error)
}

type Middleware struct {
	authServic MiddlewareContract
}

func NewMiddleware(authService MiddlewareContract) *Middleware {
	return &Middleware{
		authServic: authService,
	}
}

func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization heaader format"})
			return
		}

		tokenString := parts[1]
		claims, err := m.authServic.ValidateToken(tokenString)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
			return
		}

		// Suntikkan UserID ke context Go request saat ini
		ctx := context.WithValue(r.Context(), UserIDContextKey, claims.UserID)

		next(w, r.WithContext(ctx))
	}
}

// Helper get userID from Context
func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int64)
	return userID, ok
}

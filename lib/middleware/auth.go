package middleware

import (
	"net/http"
	"strings"

	"github.com/ariesmaulana/payroll/internal/jwtutil"
	"github.com/ariesmaulana/payroll/lib/contextutil"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims, err := jwtutil.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Inject user info ke context
		ctx := contextutil.WithUser(r.Context(), &contextutil.AuthUser{
			Id:       claims.UserID,
			Username: claims.Username,
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

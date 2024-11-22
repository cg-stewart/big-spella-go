package auth

import (
	"context"
	"net/http"
	"strings"
)

const (
	UserContextKey contextKey = "user"
)

// Middleware creates a new middleware handler for authentication
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate token
		user, err := s.ValidateToken(parts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth creates a middleware that requires authentication
func (s *Service) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserContextKey)
		if user == nil {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePremium creates a middleware that requires premium subscription
func (s *Service) RequirePremium(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(UserContextKey).(*User)
		if !user.IsPremium {
			http.Error(w, "premium subscription required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUser retrieves the user from the context
func GetUser(ctx context.Context) *User {
	user, _ := ctx.Value(UserContextKey).(*User)
	return user
}

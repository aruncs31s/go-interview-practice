package middleware

import (
	"context"
	"net/http"
	"strings"

	"web-ui/internal/services"
	"web-ui/internal/utils"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// UserContextKey is the key for user data in context
	UserContextKey ContextKey = "user"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth middleware that requires valid authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		token := utils.ExtractTokenFromHeader(authHeader)
		if token == "" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token and get user
		user, err := m.authService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		r = r.WithContext(ctx)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware that optionally validates authentication
// If token is present, it validates it and adds user to context
// If token is not present or invalid, it continues without user
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			token := utils.ExtractTokenFromHeader(authHeader)
			if token != "" {
				// Try to validate token
				user, err := m.authService.ValidateToken(token)
				if err == nil {
					// Add user to context if validation succeeds
					ctx := context.WithValue(r.Context(), UserContextKey, user)
					r = r.WithContext(ctx)
				}
			}
		}

		// Call next handler regardless of authentication status
		next.ServeHTTP(w, r)
	})
}

// CORS middleware to handle Cross-Origin Resource Sharing
func (m *AuthMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple logging - in production, use a proper logging library
		println("[", r.Method, "]", r.URL.Path, "from", r.RemoteAddr)

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

// JSONResponseMiddleware sets JSON content type for API routes
func JSONResponseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set JSON content type for API routes
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}

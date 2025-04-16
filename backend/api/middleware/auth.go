package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.uber.org/zap"
)

// AuthMiddleware verifies JWT tokens from Google Sign-In
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())

		// Get authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn("Missing Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract token
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Warn("Invalid Authorization header format")
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		// Verify JWT token and extract user ID
		authService := services.GetAuthService()
		userID, err := authService.VerifyJWT(token)
		if err != nil {
			log.Warn("Invalid JWT token", zap.Error(err))
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		log.Debug("User authenticated", zap.String("user_id", userID))
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// backend/handlers/auth.go
package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct{}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// generateStateToken generates a random state token for OAuth flow
func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// HandleGoogleLogin processes Google Sign-In requests
func (h *AuthHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Handling Google login request")

	authService := services.GetAuthService()

	// Check if it's an initial request or a callback with a code
	code := r.URL.Query().Get("code")
	if code == "" {
		// This is the initial request, redirect to Google OAuth
		state, err := generateStateToken()
		if err != nil {
			log.Error("Failed to generate state token", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Store state in a cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
		})

		// Get the auth URL
		authURL := authService.GetAuthURL(state)

		// Return the URL to the frontend
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"auth_url": authURL,
		})
		return
	}

	// This is a callback request with a code
	log.Debug("Processing OAuth callback", zap.String("code_len",
		"redacted"))

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		log.Error("Missing state cookie", zap.Error(err))
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	stateParam := r.URL.Query().Get("state")
	if stateParam == "" || stateParam != stateCookie.Value {
		log.Error("Invalid state parameter")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange the code for a token
	token, err := authService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		log.Error("Failed to exchange code for token", zap.Error(err))
		http.Error(w, "Failed to exchange code for token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	userInfo, err := authService.GetUserInfo(r.Context(), token)
	if err != nil {
		log.Error("Failed to get user info", zap.Error(err))
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}

	// Create or update user in database
	userID, err := authService.CreateOrUpdateUser(r.Context(), userInfo)
	if err != nil {
		log.Error("Failed to process user info", zap.Error(err))
		http.Error(w, "Failed to process user info", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwtToken, err := authService.GenerateJWT(userID, userInfo.Email)
	if err != nil {
		log.Error("Failed to generate JWT token", zap.Error(err))
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	log.Info("Google Sign-In successful", zap.String("user_id", userID))

	// Return JWT token and user info
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": jwtToken,
		"user": map[string]interface{}{
			"id":      userID,
			"email":   userInfo.Email,
			"name":    userInfo.Name,
			"picture": userInfo.Picture,
		},
	})
}

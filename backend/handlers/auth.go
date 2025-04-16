// backend/handlers/auth.go
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related API requests
// This includes user login, registration, and token validation
type AuthHandler struct {
	JWTSecret string // Secret key used for signing JWT tokens
}

// NewAuthHandler creates a new authentication handler with the provided dependencies
// It initializes the handler with the authentication service and JWT secret
func NewAuthHandler(jwtSecret string) *AuthHandler {
	return &AuthHandler{
		JWTSecret: jwtSecret,
	}
}

// LoginRequest represents the request body for user login
// This contains the required fields for user authentication
type LoginRequest struct {
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // User's password (will be verified against hashed version)
}

// GoogleLoginRequest represents the request body for Google OAuth login
// This contains the ID token received from Google's authentication process
type GoogleLoginRequest struct {
	IDToken string `json:"id_token"` // Google-provided ID token from OAuth flow
}

// RegisterRequest represents the request body for user registration
// This contains all the required fields to create a new user account
type RegisterRequest struct {
	Name     string `json:"name"`     // User's full name
	Email    string `json:"email"`    // User's email address
	Password string `json:"password"` // User's password (will be hashed before storage)
}

// LoginResponse represents the response body for successful login
// This contains the authentication token and user information
type LoginResponse struct {
	Token     string         `json:"token"`      // JWT token for authentication
	User      *services.User `json:"user"`       // User information
	ExpiresAt int64          `json:"expires_at"` // Token expiration timestamp
}

// Login handles user authentication with email and password
// This endpoint validates credentials and returns a JWT token if successful
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate email and password
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Email/password login not implemented - only Google Sign-in is supported
	http.Error(w, "Email/password login not implemented. Please use Google Sign-in.", http.StatusNotImplemented)
	return
}

// GoogleLogin handles authentication with Google OAuth
// This endpoint verifies the Google ID token and creates/updates the user
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate ID token
	if req.IDToken == "" {
		http.Error(w, "ID token is required", http.StatusBadRequest)
		return
	}

	// Google login not implemented
	http.Error(w, "Google login not implemented", http.StatusNotImplemented)
	return
}

// Register handles new user registration
// This endpoint creates a new user in the database with the provided information
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request fields
	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "Name, email, and password are required", http.StatusBadRequest)
		return
	}

	// User registration not implemented - only Google Sign-in is supported
	http.Error(w, "User registration not implemented. Please use Google Sign-in.", http.StatusNotImplemented)
	return
}

// AuthMiddleware creates middleware for JWT authentication
// This middleware validates JWT tokens and sets user information in the request context
func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		// Remove 'Bearer ' prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// Parse and validate token
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(h.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			logger.FromContext(r.Context()).Warn("Invalid authentication token", zap.Error(err))
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims - no need to type assert since we're using ParseWithClaims
		// Get user ID from claims
		userID, ok := claims["user_id"].(string)
		if !ok {
			http.Error(w, "Unauthorized: Invalid user ID", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), "userID", userID)

		// Also add user role if available
		if role, ok := claims["role"].(string); ok {
			ctx = context.WithValue(ctx, "userRole", role)
		}

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
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

		// Store state token in session or context for later verification
		// For simplicity, we're using a cookie here, but a more secure method would be recommended
		http.SetCookie(w, &http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			MaxAge:   int(time.Now().Add(10 * time.Minute).Unix()),
		})

		// Redirect to Google OAuth consent page
		authURL := authService.GetAuthURL(state)
		http.Redirect(w, r, authURL, http.StatusFound)
		return
	}

	// This is a callback with code, exchange it for token
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		log.Error("Failed to get state token from cookie", zap.Error(err))
		http.Error(w, "Failed to verify state token", http.StatusBadRequest)
		return
	}

	// Verify state token to prevent CSRF
	state := r.URL.Query().Get("state")
	if state == "" || state != stateCookie.Value {
		log.Error("Invalid state token", zap.String("received", state), zap.String("expected", stateCookie.Value))
		http.Error(w, "Invalid state token", http.StatusBadRequest)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})

	// Exchange code for token
	token, err := authService.ExchangeCodeForToken(r.Context(), code)
	if err != nil {
		log.Error("Failed to exchange code for token", zap.Error(err))
		http.Error(w, "Failed to exchange code", http.StatusInternalServerError)
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

// HandleTokenExchange exchanges a Google token for a backend JWT token
func (h *AuthHandler) HandleTokenExchange(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log.Info("Handling token exchange request")

	// Parse request body
	var requestBody struct {
		Token     string `json:"token"`
		TokenType string `json:"token_type"` // Optional - can be "access_token" or "id_token", defaults to "access_token"
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Error("Failed to parse request body", zap.Error(err))
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if requestBody.Token == "" {
		log.Warn("Missing token in request")
		http.Error(w, "Token is required", http.StatusBadRequest)
		return
	}

	// Default to access_token if not specified
	if requestBody.TokenType == "" {
		requestBody.TokenType = "access_token"
	}

	log.Debug("Received token for exchange",
		zap.String("token_type", requestBody.TokenType),
		zap.String("token_prefix", requestBody.Token[:min(10, len(requestBody.Token))]+"..."))

	// Determine endpoint based on token type
	endpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	var authHeader string

	if requestBody.TokenType == "id_token" {
		// For ID tokens, we need to verify with Google's token info endpoint
		endpoint = "https://oauth2.googleapis.com/tokeninfo?id_token=" + requestBody.Token
		// No auth header needed for ID token verification
		authHeader = ""
	} else {
		// Standard access token verification
		authHeader = "Bearer " + requestBody.Token
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var req *http.Request
	var err error

	if requestBody.TokenType == "id_token" {
		// For ID token, use GET without body
		req, err = http.NewRequest("GET", endpoint, nil)
	} else {
		// For access token, use GET with Authorization header
		req, err = http.NewRequest("GET", endpoint, nil)
		req.Header.Add("Authorization", authHeader)
	}

	if err != nil {
		log.Error("Failed to create request", zap.Error(err))
		http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send verification request", zap.Error(err))
		http.Error(w, "Failed to verify token: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Warn("Invalid Google token",
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(bodyBytes)))

		// Provide more detailed error message based on response
		errorMsg := fmt.Sprintf("Invalid token: Google API responded with status %s", resp.Status)
		if len(bodyBytes) > 0 {
			errorMsg += fmt.Sprintf(" - Details: %s", string(bodyBytes))
		}

		http.Error(w, errorMsg, http.StatusUnauthorized)
		return
	}

	// Parse user info
	var userInfo services.GoogleUserInfo
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", zap.Error(err))
		http.Error(w, "Failed to read user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Debug("Google userinfo response", zap.String("body", string(bodyBytes)))

	if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
		log.Error("Failed to parse user info", zap.Error(err), zap.String("body", string(bodyBytes)))
		http.Error(w, "Failed to process user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ID token response has slightly different field names than userinfo endpoint
	if requestBody.TokenType == "id_token" {
		// If using ID token and sub exists but ID doesn't, copy sub to ID
		if userInfo.ID == "" && userInfo.Sub != "" {
			userInfo.ID = userInfo.Sub
		}

		// Handle email verification status
		if userInfo.Email == "" {
			userInfo.Email = userInfo.EmailFromIDToken
		}
	}

	if userInfo.ID == "" || userInfo.Email == "" {
		log.Error("Incomplete user info from Google", zap.Any("userInfo", userInfo))
		http.Error(w, "Incomplete user info received from Google", http.StatusInternalServerError)
		return
	}

	// Get auth service
	authService := services.GetAuthService()

	// Create or update user
	userID, err := authService.CreateOrUpdateUser(r.Context(), &userInfo)
	if err != nil {
		log.Error("Failed to process user", zap.Error(err))
		http.Error(w, "Failed to process user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwtToken, err := authService.GenerateJWT(userID, userInfo.Email)
	if err != nil {
		log.Error("Failed to generate JWT", zap.Error(err))
		http.Error(w, "Failed to generate token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return JWT token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwtToken,
	})
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

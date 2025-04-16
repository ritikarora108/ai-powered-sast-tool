// backend/services/auth.go
package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	sqlcdb "github.com/ritikarora108/ai-powered-sast-tool/backend/db/sqlc"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// User represents information about a user
type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"picture"`
}

// GoogleUserInfo represents information from Google's userinfo endpoint
type GoogleUserInfo struct {
	ID               string `json:"id"`             // Standard user ID
	Sub              string `json:"sub"`            // Subject ID from ID token
	Email            string `json:"email"`          // Email from userinfo
	EmailFromIDToken string `json:"email_verified"` // Email field from ID token response
	VerifiedEmail    bool   `json:"verified_email"` // Email verification status
	Name             string `json:"name"`           // User's name
	GivenName        string `json:"given_name"`     // First name
	FamilyName       string `json:"family_name"`    // Last name
	Picture          string `json:"picture"`        // Profile picture URL
	Locale           string `json:"locale"`         // User's locale
}

// GoogleTokenResponse represents the response from Google's token endpoint
type GoogleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// AuthService handles authentication-related functions
type AuthService struct {
	config      *oauth2.Config
	dbConn      *db.Queries
	sqlcQueries *sqlcdb.Queries
}

var authService *AuthService

// Our DB connection variables for the auth service
var (
	sqlDB *sql.DB
)

// InitAuthService initializes the auth service with database connections
func InitAuthService(dbQueries *db.Queries) {
	// We can access the underlying SQL DB connection
	if dbQueries != nil {
		sqlDB = dbQueries.GetDB()
	}

	// Initialize OAuth2 config
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	// If env variables are missing, log a warning
	if googleClientID == "" || googleClientSecret == "" {
		logger.Warn("Google OAuth credentials not found in environment variables")
	}

	config := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Initialize SQLC queries
	var sqlcQueriesInstance *sqlcdb.Queries
	// We can't use the SQL DB directly with SQLC as they have different interfaces
	// For now, we'll just use the existing database interface

	authService = &AuthService{
		config:      config,
		dbConn:      dbQueries,
		sqlcQueries: sqlcQueriesInstance,
	}

	logger.Info("Auth service initialized")
}

// GetAuthService returns the auth service instance
func GetAuthService() *AuthService {
	if authService == nil {
		logger.Warn("Auth service not initialized, call InitAuthService first")
		return &AuthService{}
	}
	return authService
}

// GetAuthURL returns the Google OAuth URL for authentication
func (s *AuthService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state)
}

// ExchangeCodeForToken exchanges OAuth code for token
func (s *AuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.config.Exchange(ctx, code)
}

// ExchangeCodeForToken exchanges an authorization code for an access token (legacy function)
func ExchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return "", fmt.Errorf("missing required environment variables")
	}

	// Create the request to exchange the auth code for an access token
	tokenURL := "https://oauth2.googleapis.com/token"
	req, err := http.NewRequest("POST", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %v", err)
	}

	q := req.URL.Query()
	q.Add("grant_type", "authorization_code")
	q.Add("code", code)
	q.Add("client_id", clientID)
	q.Add("client_secret", clientSecret)
	q.Add("redirect_uri", redirectURI)
	req.URL.RawQuery = q.Encode()

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %s", body)
	}

	// Parse the response
	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %v", err)
	}

	return tokenResp.AccessToken, nil
}

// GetUserInfo gets user info from Google using OAuth2 token
func (s *AuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*GoogleUserInfo, error) {
	client := s.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		logger.Error("Failed to get user info from Google", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error("Non-OK response from Google userinfo",
			zap.Int("status_code", resp.StatusCode))
		return nil, fmt.Errorf("non-OK response from Google: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, err
	}

	var userInfo GoogleUserInfo
	if err := json.Unmarshal(data, &userInfo); err != nil {
		logger.Error("Failed to unmarshal user info", zap.Error(err))
		return nil, err
	}

	return &userInfo, nil
}

// GetUserInfo retrieves user information from Google using the access token (legacy function)
func GetUserInfo(token string) (*User, error) {
	// Create the request to get user info
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user info request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed: %s", body)
	}

	// Parse the response
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %v", err)
	}

	return &user, nil
}

// CreateOrUpdateUser creates or updates a user based on Google info
func (s *AuthService) CreateOrUpdateUser(ctx context.Context, userInfo *GoogleUserInfo) (string, error) {
	// For now, we'll use a simpler implementation that doesn't use SQLC
	// This avoids complex SQL driver compatibility issues

	if s.dbConn == nil || s.dbConn.GetDB() == nil {
		return "", errors.New("database connection not initialized")
	}

	db := s.dbConn.GetDB()

	// Check if user exists by Google ID or email
	var userID string
	var exists bool

	err := db.QueryRowContext(ctx,
		"SELECT id, true FROM users WHERE email = $1 OR google_id = $2 LIMIT 1",
		userInfo.Email, userInfo.ID).Scan(&userID, &exists)

	if err != nil && err != sql.ErrNoRows {
		logger.Error("Database error checking for existing user", zap.Error(err))
		return "", err
	}

	if err == sql.ErrNoRows || !exists {
		// User doesn't exist, create them
		logger.Info("Creating new user", zap.String("email", userInfo.Email))

		// Insert user into database
		err := db.QueryRowContext(ctx,
			"INSERT INTO users (email, name, google_id, avatar_url) VALUES ($1, $2, $3, $4) RETURNING id",
			userInfo.Email, userInfo.Name, userInfo.ID, userInfo.Picture).Scan(&userID)

		if err != nil {
			logger.Error("Failed to create user", zap.Error(err))
			return "", err
		}

		logger.Info("User created successfully", zap.String("user_id", userID))
		return userID, nil
	}

	// User exists, update them
	logger.Info("Updating existing user", zap.String("email", userInfo.Email))

	_, err = db.ExecContext(ctx,
		"UPDATE users SET name = $1, google_id = $2, avatar_url = $3, updated_at = NOW() WHERE id = $4",
		userInfo.Name, userInfo.ID, userInfo.Picture, userID)

	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return "", err
	}

	logger.Info("User updated successfully", zap.String("user_id", userID))
	return userID, nil
}

// ProcessUserInfo processes user information by creating or updating the user in the database
func ProcessUserInfo(user *User) error {
	// This is kept for backward compatibility
	// In a real application, you would fully implement the database operations
	return nil
}

// GenerateJWT generates a JWT token for the user
func (s *AuthService) GenerateJWT(userID, email string) (string, error) {
	// Get the JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production" // Fallback secret
		logger.Warn("Using default JWT secret, consider setting JWT_SECRET environment variable")
	}

	// Create claims with user information
	expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-powered-sast-tool",
			Subject:   userID,
		},
	}

	// Create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", err
	}

	return signedToken, nil
}

// GenerateSessionToken generates a JWT token for the authenticated user
func GenerateSessionToken(user *User) (string, error) {
	// This function is kept for backward compatibility

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-key" // Default for development, should be properly set in production
	}

	// Set expiration time
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	// Create claims with user data
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-powered-sast-tool",
			Subject:   user.ID,
		},
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

// VerifyJWT verifies a JWT token and returns the user ID
func (s *AuthService) VerifyJWT(tokenString string) (string, error) {
	// Get the JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-key-change-in-production" // Fallback secret
	}

	// Parse and validate the token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		logger.Warn("Failed to parse JWT token", zap.Error(err))
		return "", err
	}

	if !token.Valid {
		logger.Warn("Invalid JWT token")
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

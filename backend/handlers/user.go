package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
)

// User represents a user profile in the system
// This struct defines the user information that is returned to the client
// It contains only the public, non-sensitive information about a user
type User struct {
	ID        string    `json:"id"`         // Unique identifier for the user
	Email     string    `json:"email"`      // User's email address
	Name      string    `json:"name"`       // User's full name
	AvatarURL string    `json:"avatar_url"` // URL to the user's profile picture
	CreatedAt time.Time `json:"created_at"` // When the user account was created
}

// HandleGetUserProfile returns the profile of the authenticated user
// This handler retrieves user information from the database
// It requires authentication and expects the user ID to be in the request context
func HandleGetUserProfile(w http.ResponseWriter, r *http.Request, dbQueries *db.Queries) {
	log := logger.FromContext(r.Context())

	// Get user ID from context (set by auth middleware)
	// The auth middleware must have run before this handler to set the user ID
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Warn("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Verify that database connection is available
	// The dbQueries parameter is expected to contain an initialized database connection
	if dbQueries == nil || dbQueries.GetDB() == nil {
		log.Error("Database connection not initialized")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Debug("Getting user profile", zap.String("user_id", userID))

	// Query user profile from database using the user ID
	// This retrieves the user's basic information for display
	var user User
	err := dbQueries.GetDB().QueryRowContext(
		r.Context(),
		"SELECT id, email, name, avatar_url, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)

	// Handle potential database errors
	if err != nil {
		if err == sql.ErrNoRows {
			// No user found with the provided ID
			log.Warn("User not found", zap.String("user_id", userID))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Other database errors
		log.Error("Database error", zap.String("user_id", userID), zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return user profile as JSON response
	// This includes all the fields defined in the User struct
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

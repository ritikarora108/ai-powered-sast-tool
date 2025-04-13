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

// User represents a user profile
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// HandleGetUserProfile returns the profile of the authenticated user
func HandleGetUserProfile(w http.ResponseWriter, r *http.Request, dbQueries *db.Queries) {
	log := logger.FromContext(r.Context())

	// Get user ID from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Warn("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database using direct SQL
	if dbQueries == nil || dbQueries.GetDB() == nil {
		log.Error("Database connection not initialized")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Debug("Getting user profile", zap.String("user_id", userID))

	// Query user from database
	var user User
	err := dbQueries.GetDB().QueryRowContext(
		r.Context(),
		"SELECT id, email, name, avatar_url, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Warn("User not found", zap.String("user_id", userID))
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		log.Error("Database error", zap.String("user_id", userID), zap.Error(err))
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return user profile
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

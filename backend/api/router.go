package api

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/api/middleware"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/handlers"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.temporal.io/sdk/client"
)

func NewRouter(temporalClient client.Client, dbQueries *db.Queries) *chi.Mux {
	// Initialize our logger
	logger.Init()

	router := chi.NewRouter()

	// Add all middleware first
	router.Use(middleware.RequestLogger) // Use our custom logger instead of chi's
	router.Use(chimiddleware.Recoverer)
	router.Use(chimiddleware.Timeout(60 * time.Second))
	router.Use(chimiddleware.RequestID) // Generate request IDs

	// Setup CORS
	frontendURL := os.Getenv("FRONTEND_URL")
	corsOrigins := []string{"*"} // Default to allow all origins

	// If FRONTEND_URL is set, use that as the allowed origin
	if frontendURL != "" {
		logger.Info("Setting CORS allowed origin to " + frontendURL)
		corsOrigins = []string{frontendURL}

		// Add additional origins if needed (comma-separated list)
		additionalOrigins := os.Getenv("ADDITIONAL_CORS_ORIGINS")
		if additionalOrigins != "" {
			origins := strings.Split(additionalOrigins, ",")
			corsOrigins = append(corsOrigins, origins...)
		}
	}

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not rarely used
	})
	router.Use(corsMiddleware.Handler)

	// Health check route (no auth needed)
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.FromContext(r.Context()).Debug("Health check called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize auth service with database connection
	services.InitAuthService(dbQueries)

	// Create services
	githubService := services.NewGitHubService(dbQueries)
	scannerService := services.NewScannerService(githubService)
	openAIService := services.NewOpenAIService()

	// Log service initialization
	logger.Info("Services initialized successfully")

	// Auth routes (for login)
	router.Route("/auth", func(r chi.Router) {
		authHandler := handlers.NewAuthHandler()
		r.Get("/google", authHandler.HandleGoogleLogin)
		r.Get("/google/callback", authHandler.HandleGoogleLogin)
	})

	// Public scanning endpoint - no auth required
	repositoryHandler := handlers.NewRepositoryHandler(githubService, scannerService, openAIService, temporalClient)
	router.Post("/scan", repositoryHandler.ScanPublicRepository)
	router.Get("/scan/{id}/status", repositoryHandler.GetScanStatus)
	router.Get("/scan/{id}/results", repositoryHandler.GetScanResults)
	router.Get("/scan/{id}/debug", repositoryHandler.DebugWorkflow)

	// Protected API routes
	router.Route("/api", func(r chi.Router) {
		// Apply authentication middleware to all /api routes
		r.Use(middleware.AuthMiddleware)

		// Repository routes
		r.Route("/repositories", func(r chi.Router) {
			r.Post("/", repositoryHandler.CreateRepository)
			r.Get("/", repositoryHandler.ListRepositories)
			r.Get("/{id}", repositoryHandler.GetRepository)
			r.Post("/{id}/scan", repositoryHandler.ScanRepository)
			r.Get("/{id}/vulnerabilities", repositoryHandler.GetVulnerabilities)
		})

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				// Implement user profile endpoint
				// This will be automatically protected by the auth middleware
				handlers.HandleGetUserProfile(w, r, dbQueries)
			})
		})
	})

	logger.Info("Router initialized with all routes")
	return router
}

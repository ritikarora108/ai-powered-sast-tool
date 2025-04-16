package api

import (
	"encoding/json"
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
	"go.uber.org/zap"
)

// NewRouter creates and configures the application's HTTP router
// This function sets up all API routes, middleware, and handler dependencies
// It is the central point for defining the API structure and route handlers
func NewRouter(temporalClient client.Client, dbQueries *db.Queries) *chi.Mux {
	// Initialize our logger for structured logging
	logger.Init()

	// Create a new Chi router - Chi is a lightweight, idiomatic router for Go HTTP services
	router := chi.NewRouter()

	// Set up global middleware that will apply to all routes
	// Middleware is executed in the order it's added
	router.Use(middleware.RequestLogger)                // Custom logger for HTTP requests
	router.Use(chimiddleware.Recoverer)                 // Recover from panics without crashing the server
	router.Use(chimiddleware.Timeout(60 * time.Second)) // Set a timeout for all requests
	router.Use(chimiddleware.RequestID)                 // Generate unique request IDs for tracking

	// Add custom middleware for graceful error handling
	// This catches any panics in handlers and returns a 500 error instead of crashing
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log := logger.FromContext(r.Context())
					log.Error("Panic in handler",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("method", r.Method),
					)

					w.WriteHeader(http.StatusInternalServerError)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{
						"error": "Internal server error",
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	})

	// Set up Cross-Origin Resource Sharing (CORS) configuration
	// This allows controlled access to the API from different domains
	frontendURL := os.Getenv("FRONTEND_URL")
	corsOrigins := []string{"*"} // Default to allow all origins

	// If FRONTEND_URL environment variable is set, use that as the allowed origin
	if frontendURL != "" {
		logger.Info("Setting CORS allowed origin to " + frontendURL)
		corsOrigins = []string{frontendURL}

		// Add additional origins if needed (comma-separated list from environment)
		additionalOrigins := os.Getenv("ADDITIONAL_CORS_ORIGINS")
		if additionalOrigins != "" {
			origins := strings.Split(additionalOrigins, ",")
			corsOrigins = append(corsOrigins, origins...)
		}

		// Always include localhost origins for development environments
		corsOrigins = append(corsOrigins,
			"http://localhost:3000",
			"http://127.0.0.1:3000",
			"http://localhost:8080",
			"http://127.0.0.1:8080")
	}

	// Log CORS origins for debugging purposes
	logger.Info("CORS origins configured", zap.Strings("origins", corsOrigins))

	// Create and apply the CORS middleware with our configuration
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   corsOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value in seconds for preflight cache
	})
	router.Use(corsMiddleware.Handler)

	// Health check endpoint for monitoring and load balancers
	// This simple endpoint allows checking if the API is running
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.FromContext(r.Context()).Debug("Health check called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Initialize authentication service with database connection
	services.InitAuthService(dbQueries)

	// Create service instances
	// These services contain the business logic of the application
	githubService := services.NewGitHubService(dbQueries)
	scannerService := services.NewScannerService(githubService)
	openAIService := services.NewOpenAIService()

	// Log successful service initialization
	logger.Info("Services initialized successfully")

	// Authentication routes
	// These handle OAuth flows and token generation
	router.Route("/auth", func(r chi.Router) {
		// Create auth handler with JWT secret
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "default-secret-for-development-only"
			logger.Warn("JWT_SECRET environment variable not set, using default secret (not secure for production)")
		}

		authHandler := handlers.NewAuthHandler(jwtSecret)

		r.Get("/google", authHandler.HandleGoogleLogin)          // Initiate Google OAuth flow
		r.Get("/google/callback", authHandler.HandleGoogleLogin) // OAuth callback from Google
		r.Post("/token", authHandler.HandleTokenExchange)        // Exchange OAuth code for JWT token
	})

	// Public scanning endpoints - no authentication required
	// These allow anonymous users to scan public repositories
	repositoryHandler := handlers.NewRepositoryHandler(githubService, scannerService, openAIService, temporalClient)
	router.Post("/scan", repositoryHandler.ScanPublicRepository)       // Start a scan for a public repo
	router.Get("/scan/{id}/status", repositoryHandler.GetScanStatus)   // Check scan status by ID
	router.Get("/scan/{id}/results", repositoryHandler.GetScanResults) // Get scan results by ID
	router.Get("/scan/{id}/debug", repositoryHandler.DebugWorkflow)    // Debugging endpoint for workflows

	// Repository routes - protected by authentication
	// These endpoints manage repositories and their scans
	router.Route("/repositories", func(r chi.Router) {
		// Apply authentication middleware to all routes in this group
		r.Use(middleware.AuthMiddleware)

		r.Post("/", repositoryHandler.CreateRepository)                      // Create a new repository
		r.Get("/", repositoryHandler.ListRepositories)                       // List all repositories for current user
		r.Get("/{id}", repositoryHandler.GetRepository)                      // Get details of a specific repository
		r.Post("/{id}/scan", repositoryHandler.ScanRepository)               // Start a scan for a specific repository
		r.Get("/{id}/vulnerabilities", repositoryHandler.GetVulnerabilities) // Get vulnerabilities for a repository
	})

	// Protected API routes - general purpose endpoints that require authentication
	router.Route("/api", func(r chi.Router) {
		// Apply authentication middleware to all /api routes
		r.Use(middleware.AuthMiddleware)

		// User management routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				// Get the authenticated user's profile
				handlers.HandleGetUserProfile(w, r, dbQueries)
			})
		})
	})

	logger.Info("Router initialized with all routes")
	return router
}

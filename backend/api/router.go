package api

import (
	"log"
	"net/http"	
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/ritikarora108/ai-powered-sast-tool/backend/handlers"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/services"
	"go.temporal.io/sdk/client"
)

func NewRouter(temporalClient client.Client) *chi.Mux {
	router := chi.NewRouter()
	
	// Add all middleware first
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	
	// Health check route
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	githubService := services.NewGitHubService()
	scannerService := services.NewScannerService(githubService)
	openAIService := services.NewOpenAIService()
	temporalClient, err := client.NewLazyClient(client.Options{})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}

	// Set up API routes
	router.Route("/api", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("API Endpoint"))
		})
		
		// Repository routes
		r.Route("/repositories", func(r chi.Router) {
			repositoryHandler := handlers.NewRepositoryHandler(githubService, scannerService, openAIService, temporalClient)
			r.Post("/", repositoryHandler.CreateRepository)
			r.Get("/", repositoryHandler.ListRepositories)
			r.Get("/{id}", repositoryHandler.GetRepository)
			r.Post("/{id}/scan", repositoryHandler.ScanRepository)
			r.Get("/{id}/vulnerabilities", repositoryHandler.GetVulnerabilities)
		})
	})

	return router
}
package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/api"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/db"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

// startScanWorker initializes and starts a Temporal worker to process tasks from the SCAN_TASK_QUEUE
// This worker will execute the scan workflows and activities asynchronously
func startScanWorker(c client.Client) error {
	logger.Info("Creating Temporal worker for SCAN_TASK_QUEUE")

	// Create worker options with concurrency limits to prevent overloading the system
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:     5,  // Limit concurrent activities
		MaxConcurrentWorkflowTaskExecutionSize: 10, // Limit concurrent workflows
	}

	// Create a new worker connected to the SCAN_TASK_QUEUE
	w := worker.New(c, "SCAN_TASK_QUEUE", workerOptions)

	// Register workflow and activities with the worker
	// These define what code will be executed when tasks are received
	logger.Info("Registering workflow and activities")
	w.RegisterWorkflow(temporal.ScanWorkflow)
	w.RegisterActivity(temporal.CloneRepositoryActivity)
	w.RegisterActivity(temporal.ScanRepositoryActivity)

	// Start the worker (non-blocking)
	// This will run in the background listening for tasks
	logger.Info("Starting Temporal worker")
	return w.Start()
}

// main is the entry point for the application
// It initializes all components and starts the HTTP server
func main() {
	// Load environment variables from .env file
	// This makes local development easier by not requiring environment variables to be set system-wide
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Initialize the logger for structured logging throughout the application
	logger.Init()
	defer logger.Sync() // Ensure all buffered logs are written on exit

	logger.Info("Starting AI-powered SAST tool backend")

	// Connect to PostgreSQL database - extract connection parameters from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	logger.Info("Connecting to PostgreSQL database",
		zap.String("host", dbHost),
		zap.String("port", dbPort),
		zap.String("database", dbName),
		zap.String("user", dbUser))

	// Build the PostgreSQL connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open a connection to the database
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
		os.Exit(1)
	}

	// Configure connection pool settings for optimal performance
	sqlDB.SetMaxOpenConns(25)                 // Maximum number of open connections to the database
	sqlDB.SetMaxIdleConns(5)                  // Maximum number of connections in the idle connection pool
	sqlDB.SetConnMaxLifetime(time.Minute * 5) // Maximum amount of time a connection may be reused

	// Test database connection with a timeout to ensure it's working
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add retry logic for database connection in case of temporary issues
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err = sqlDB.PingContext(ctx)
		if err == nil {
			logger.Info("Successfully connected to PostgreSQL database")
			break
		}

		logger.Warn("Failed to ping database, retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxRetries))

		if i < maxRetries-1 {
			time.Sleep(time.Second * 2)
		}
	}

	// Continue even if database connection failed
	// This allows the application to start without a database
	// and will use in-memory storage or fail gracefully for database operations
	if err != nil {
		logger.Error("Failed to connect to database after multiple attempts", zap.Error(err))
		logger.Warn("Continuing without database connection - some features may not work")
	}

	// Set the global database connection for use throughout the application
	db.SetGlobalDB(sqlDB)
	defer sqlDB.Close()

	// Initialize database queries wrapper
	dbQueries := db.NewQueries()
	defer dbQueries.Close()

	// Initialize Temporal client for workflow orchestration
	// Temporal is used for managing long-running scan workflows
	logger.Info("Initializing Temporal client")
	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort: os.Getenv("TEMPORAL_HOST"),
	})
	if err != nil {
		logger.Fatal("Unable to create Temporal client", zap.Error(err))
	}
	defer temporalClient.Close()

	// Start Temporal worker for scan workflows
	// This worker will execute the repository scanning tasks asynchronously
	logger.Info("Starting Temporal worker for scan workflows")
	err = startScanWorker(temporalClient)
	if err != nil {
		logger.Fatal("Unable to start Temporal worker", zap.Error(err))
	}

	// Create router with the temporal client and database
	// This sets up all the HTTP API endpoints
	logger.Info("Initializing API router")
	router := api.NewRouter(temporalClient, dbQueries)

	// Configure the HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not specified
	}
	serverAddr := ":" + port

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Setup server context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	// This ensures open connections are properly closed when the server is terminated
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-sig
		logger.Info("Received shutdown signal", zap.String("signal", s.String()))

		// Shutdown signal with grace period of 30 seconds
		// This allows ongoing requests to complete
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("Graceful shutdown timed out... forcing exit")
			}
		}()

		// Trigger graceful shutdown of the HTTP server
		logger.Info("Shutting down server gracefully")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal("Server shutdown failed", zap.Error(err))
		}
		serverStopCtx()
	}()

	// Start the HTTP server
	logger.Info("Server starting", zap.String("address", serverAddr))
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server error", zap.Error(err))
	}

	// Wait for server context to be stopped
	// This blocks until the server has completely shut down
	<-serverCtx.Done()
	logger.Info("Server stopped gracefully")
}

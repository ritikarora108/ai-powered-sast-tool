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
func startScanWorker(c client.Client) error {
	logger.Info("Creating Temporal worker for SCAN_TASK_QUEUE")

	// Create worker options
	workerOptions := worker.Options{
		MaxConcurrentActivityExecutionSize:     5,  // Limit concurrent activities
		MaxConcurrentWorkflowTaskExecutionSize: 10, // Limit concurrent workflows
	}

	// Create a new worker
	w := worker.New(c, "SCAN_TASK_QUEUE", workerOptions)

	// Register workflow and activities
	logger.Info("Registering workflow and activities")
	w.RegisterWorkflow(temporal.ScanWorkflow)
	w.RegisterActivity(temporal.CloneRepositoryActivity)
	w.RegisterActivity(temporal.ScanRepositoryActivity)

	// Start the worker (non-blocking)
	logger.Info("Starting Temporal worker")
	return w.Start()
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Initialize logger
	logger.Init()
	defer logger.Sync() // Ensure all buffered logs are written on exit

	logger.Info("Starting AI-powered SAST tool backend")

	// Initialize Temporal client
	logger.Info("Initializing Temporal client")
	temporalClient, err := client.NewLazyClient(client.Options{
		HostPort: os.Getenv("TEMPORAL_HOST"),
	})
	if err != nil {
		logger.Fatal("Unable to create Temporal client", zap.Error(err))
	}
	defer temporalClient.Close()

	// Start Temporal worker for scan workflows
	logger.Info("Starting Temporal worker for scan workflows")
	err = startScanWorker(temporalClient)
	if err != nil {
		logger.Fatal("Unable to start Temporal worker", zap.Error(err))
	}

	// Connect to PostgreSQL database
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

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Try to connect to the database
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
		os.Exit(1)
	}

	// Test connection
	// err = sqlDB.Ping()
	// if err != nil {
	// 	logger.Fatal("Failed to ping database", zap.Error(err))
	// 	os.Exit(1)
	// }

	logger.Info("Successfully connected to PostgreSQL database")
	defer sqlDB.Close()

	// Initialize database queries
	dbQueries := db.NewQueries()
	dbQueries.SetDB(sqlDB)
	defer dbQueries.Close()

	// Create router with the temporal client and database
	logger.Info("Initializing API router")
	router := api.NewRouter(temporalClient, dbQueries)

	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	serverAddr := ":" + port

	server := &http.Server{
		Addr:    serverAddr,
		Handler: router,
	}

	// Server shutdown context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		s := <-sig
		logger.Info("Received shutdown signal", zap.String("signal", s.String()))

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("Graceful shutdown timed out... forcing exit")
			}
		}()

		// Trigger graceful shutdown
		logger.Info("Shutting down server gracefully")
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			logger.Fatal("Server shutdown failed", zap.Error(err))
		}
		serverStopCtx()
	}()

	// Start the server
	logger.Info("Server starting", zap.String("address", serverAddr))
	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server error", zap.Error(err))
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	logger.Info("Server stopped gracefully")
}

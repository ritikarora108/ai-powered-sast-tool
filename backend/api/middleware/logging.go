// backend/api/middleware/logging.go
package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/ritikarora108/ai-powered-sast-tool/backend/internal/logger"
	"go.uber.org/zap"
)

// RequestLogger is a middleware that logs requests using zap
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID if not present
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			r.Header.Set("X-Request-ID", requestID)
		}

		// Create a response writer wrapper to capture status code and response size
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Get start time
		start := time.Now()

		// Initialize request-scoped logger
		log := logger.With(
			zap.String("request_id", requestID),
			zap.String("remote_addr", r.RemoteAddr),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.String("user_agent", r.UserAgent()),
		)

		// Set the request-scoped logger to the request context
		ctx := r.Context()
		ctx = logger.WithContext(ctx, log)
		r = r.WithContext(ctx)

		// Log request start
		log.Debug("Request started")

		// Process the request
		next.ServeHTTP(ww, r)

		// Calculate request duration
		duration := time.Since(start)

		// Get content length
		contentLength := ww.BytesWritten()

		// Log request completion with additional metadata
		log.Info("Request completed",
			zap.Int("status", ww.Status()),
			zap.Int("content_length", contentLength),
			zap.String("duration", duration.String()),
			zap.Duration("duration_ms", duration),
		)
	})
}

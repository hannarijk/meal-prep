package middleware

import (
	"net/http"
	"time"

	"meal-prep/shared/logging"

	"github.com/google/uuid"
)

func LoggingMiddleware(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			w.Header().Set("X-Request-ID", requestID)

			// Add to context
			ctx := logging.WithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Wrapper to capture status code
			ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			// Log incoming request
			logging.WithContext(ctx).Info("Incoming request",
				"method", r.Method,
				"uri", r.RequestURI,
				"remote_ip", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Process request
			next.ServeHTTP(ww, r)

			// Log completed request
			duration := time.Since(start)
			logging.WithContext(ctx).Info("Request completed",
				"method", r.Method,
				"uri", r.RequestURI,
				"status_code", ww.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

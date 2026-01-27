package middleware

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
)

// LoggingMiddleware creates a middleware that adds request ID to context and logs requests.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get request ID from chi middleware
		requestID := chimiddleware.GetReqID(r.Context())
		ctx := logger.ContextWithRequestID(r.Context(), requestID)

		// Wrap response writer to capture status code
		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)

		logger.Info(ctx, "request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remoteAddr", r.RemoteAddr,
			"userAgent", r.UserAgent(),
		)

		// Continue with updated context
		next.ServeHTTP(ww, r.WithContext(ctx))

		logger.Info(ctx, "request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration", time.Since(start).String(),
			"durationMs", time.Since(start).Milliseconds(),
		)
	})
}

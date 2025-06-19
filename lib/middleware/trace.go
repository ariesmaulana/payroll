package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ariesmaulana/payroll/lib/contextutil"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate new UUID for trace ID
		traceId := uuid.New().String()

		// Create a custom response writer to capture the status code
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Read request body if it's available (to capture in trace)
		var body string
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err == nil {
				// Restore the body for downstream handlers
				body = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for further use
			}
		}

		// Create a Trace object with additional request info
		trace := &contextutil.Trace{
			TraceID: traceId,
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: r.Header,
			Body:    body, // Optional, capture body content
		}

		// Add trace ID to context with the defined constant key
		ctx := contextutil.WithTrace(r.Context(), trace)
		fmt.Println(ctx)
		// Add trace ID to response headers
		w.Header().Set("X-Trace-ID", traceId)

		// Create new request with updated context
		r = r.WithContext(ctx)

		// Log the request details (including trace information)
		log.Info().
			Str("traceId", traceId).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_ip", r.RemoteAddr).
			Int("status", ww.Status()).
			Dur("latency", time.Since(start)).
			Str("user_agent", r.UserAgent()).
			Msg("Request handled")

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/meesooqa/go-lgr"
)

// Logging creates a middleware that logs every request:
// method, path, response status, duration, and client address.
// If logger is nil, slog.Default() is used.
//
// Additionally, the middleware attaches a request-scoped logger
// (enriched with a request_id) to the request's context.Context via lgr.NewContext.
// Controller handlers can retrieve it using lgr.FromContext(r.Context()),
// ensuring all their log entries are automatically tagged
// with the same request_id as the final "request completed" entry.
//
// The request_id is extracted from the X-Request-Id header
// if provided by the client or proxy; otherwise, a random one is generated.
// In either case, it is returned in the X-Request-Id response header.
func Logging(logger *slog.Logger) Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = newRequestID()
			}
			w.Header().Set("X-Request-Id", requestID)

			reqLogger := logger.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
			)
			r = r.WithContext(lgr.NewContext(r.Context(), reqLogger))

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			reqLogger.Info("request completed",
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the response status code,
// since http.ResponseWriter itself does not expose it
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// newRequestID generates a random request ID
// using crypto/rand avoiding external UUID libraries
func newRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand should practically never return an error on typical OSes;
		// as a last resort, we gracefully degrade without panicking.
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000")))
	}
	return hex.EncodeToString(b)
}

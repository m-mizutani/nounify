package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.NewString()
		ctx := r.Context()
		logger := ctxutil.Logger(ctx).With("request_id", reqID)

		ctx = ctxutil.WithLogger(ctx, logger)

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		ts := time.Now()
		next.ServeHTTP(sw, r.WithContext(ctx))
		latency := time.Since(ts)

		logger.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
			"headers", r.Header,
			"latency", latency,
		)
	})
}

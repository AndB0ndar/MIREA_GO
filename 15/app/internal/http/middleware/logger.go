package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Прокси для захвата статуса ответа
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logger, _ := zap.NewProduction()
		defer logger.Sync()

		logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", duration),
			zap.String("user_agent", r.UserAgent()),
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

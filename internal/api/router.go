package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	mw "github.com/jsi/ibs-doc-engine/internal/api/middleware"
)

type RouterConfig struct {
	CORSAllowedOrigins []string
	RateLimiter        func(http.Handler) http.Handler // optional global rate limiter
}

func NewRouter(cfg RouterConfig) *chi.Mux {
	r := chi.NewRouter()

	r.Use(mw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(structuredLogger)
	r.Use(chimw.Recoverer)

	allowedOrigins := cfg.CORSAllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	if cfg.RateLimiter != nil {
		r.Use(cfg.RateLimiter)
	}

	r.Get("/health", HealthHandler("api-gateway"))

	return r
}

func structuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", mw.GetRequestID(r.Context()),
		)
	})
}

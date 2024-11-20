package api

import (
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api/middlewares"
	v1 "github.com/ercross/payment_gateways/internal/api/v1"
	"github.com/ercross/payment_gateways/internal/kafka"
	"github.com/ercross/payment_gateways/internal/logger"
	cache "github.com/ercross/payment_gateways/internal/redis"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func NewServer(
	repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
	dstrRL *middlewares.DistributedRateLimiter,
	baseURL string,
) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(logger.RequestLogger(log))
	mux.Use(middleware.Recoverer)
	mux.Use(middlewares.CORSMiddleware(baseURL))
	mux.Use(middlewares.SecurityMiddleware)

	mux.Mount("/api/v1", v1.AddRoutes(repo, log, publisher, dstrCache, dstrRL, baseURL))
	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	return mux
}

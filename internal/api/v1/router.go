package v1

import (
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api/middlewares"
	"github.com/ercross/payment_gateways/internal/kafka"
	"github.com/ercross/payment_gateways/internal/logger"
	cache "github.com/ercross/payment_gateways/internal/redis"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const (
	withdrawalCallbackPath = "api/v1/callback/withdrawals"
	depositCallbackPath    = "api/v1/callback/deposits"
)

func AddRoutes(
	repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
	dstrRL *middlewares.DistributedRateLimiter,
	baseURL string,
) http.Handler {
	router := chi.NewRouter()

	router.Mount("/callback", callbackRoutes(repo, log, publisher, dstrCache))
	router.Mount("/", paymentsInitiationRoutes(repo, log, publisher, dstrCache, dstrRL, baseURL))

	return router
}

func callbackRoutes(repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
) http.Handler {
	router := chi.NewRouter()

	router.Put("/withdrawal/{transaction-id}", withdrawalCallbackHandler(repo, log, publisher, dstrCache))
	router.Put("/deposit/{transaction-id}", depositCallbackHandler(repo, log, publisher, dstrCache))

	return router
}

func paymentsInitiationRoutes(repo db.Repository,
	log *logger.Logger,
	publisher kafka.EventPublisher,
	dstrCache cache.DistributedCache,
	dstrRL *middlewares.DistributedRateLimiter,
	baseURL string,
) http.Handler {
	router := chi.NewRouter()
	router.Use(middlewares.Authenticate)
	router.Use(dstrRL.Middleware)

	router.Post("/withdrawal", initiateWithdrawal(repo, log, publisher, dstrCache, baseURL))
	router.Post("/deposit", handleDeposit(repo, log, publisher, dstrCache, baseURL))

	return router
}

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ercross/payment_gateways/db"
	"github.com/ercross/payment_gateways/internal/api"
	"github.com/ercross/payment_gateways/internal/api/middlewares"
	"github.com/ercross/payment_gateways/internal/kafka"
	"github.com/ercross/payment_gateways/internal/logger"
	cache "github.com/ercross/payment_gateways/internal/redis"
	"github.com/ercross/payment_gateways/internal/services"
	"strconv"

	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	ctx := context.Background()

	if err := run(ctx, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	loggerConfig := &logger.Config{
		Level:       logger.INFO,
		Destination: new(logger.ConsoleDestination),
	}
	log, err := logger.NewLogger(loggerConfig)
	if err != nil {
		return fmt.Errorf("error initialising logger: %w", err)
	}

	log.Info("logger initialized...")

	dsn := os.Getenv("DATABASE_URL")
	repo, err := db.New(dsn)
	if err != nil {
		return fmt.Errorf("error initialising database: %w", err)
	}
	log.Info("Database initialized...")

	migrationFilesDir := os.Getenv("MIGRATIONS")

	if err = repo.Migrate(migrationFilesDir, dsn); err != nil {
		return fmt.Errorf("error migrating db schema: %w", err)
	}
	log.Info("Database schema files migrated successfully...")

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		return fmt.Errorf("invalid redis db value: %w", err)
	}
	redis, err := cache.New(redisAddr, redisPassword, redisDB)
	if err != nil {
		return fmt.Errorf("error initialising redis: %w", err)
	}
	log.Info("Redis initialised...")

	dstrRL := middlewares.NewDistributedRateLimiter(redis.Client(), 3, time.Minute*1)

	kafkaBrokerAddr := os.Getenv("KAFKA_BROKER_URL")
	publisher := kafka.NewProducer(kafkaBrokerAddr)
	log.Info("Kafka initialised...")
	services.InitEncryptionKey("W-Dm='U]Pu@xk]GM")
	srv := api.NewServer(repo, log, publisher, redis, dstrRL, os.Getenv("API_URL"))

	httpServer := &http.Server{
		Addr:    net.JoinHostPort("", os.Getenv("API_PORT")),
		Handler: srv,
	}
	go func() {

		log.Info("listening on %s\n", logger.NewField("Address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Error("error listening and serving: %s\n", logger.NewField("Error", err.Error()))
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Error("error shutting down http server: %s\n", logger.NewField("Error", err.Error()))
		}
	}()
	wg.Wait()
	return nil
}

package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ercross/payment_gateways/internal/services"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"time"
)

type Redis struct {
	client *redis.Client
	ctx    context.Context
	rs     *redsync.Redsync
}

func ConstructUserIDKey(userID int) string {
	return fmt.Sprintf("user:%d", userID)
}

func ConstructTransactionIDKey(trxID int) string {
	return fmt.Sprintf("transaction:%d", trxID)
}

func New(address, password string, db int) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})

	err := services.RetryOperation(func() error {
		status := client.Ping(context.Background())
		if status.Err() != nil {
			return fmt.Errorf("failed to connect to redis: %w", status.Err())
		}
		return nil
	}, 5)
	if err != nil {
		return nil, err
	}

	pool := goredis.NewPool(client)
	return &Redis{
		client: client,
		ctx:    context.Background(),
		rs:     redsync.New(pool),
	}, nil
}

func (r *Redis) Client() *redis.Client {
	return r.client
}

func (r *Redis) Save(key string, value interface{}, expiration time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = r.client.Set(r.ctx, key, string(raw), expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *Redis) Get(ctx context.Context, key string, out interface{}) error {
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil // cache miss
		}
		return err
	}
	if err = json.Unmarshal([]byte(value), out); err != nil {
		return err
	}
	return nil
}

func (r *Redis) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *Redis) AcquireLock(ctx context.Context, key string) (*Lock, error) {
	mutex := r.rs.NewMutex(key)
	if err := mutex.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}
	return &Lock{mutex: mutex}, nil
}

func (r *Redis) ReleaseLock(lock *Lock) error {
	// todo use retry
	unlocked, err := lock.mutex.Unlock()
	if err != nil {
		return fmt.Errorf("error releasing lock: %w", err)
	}
	if !unlocked {
		return errors.New("failed to unlock lock")
	}
	return nil
}

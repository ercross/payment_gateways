package cache

import (
	"context"
	"errors"
	"github.com/go-redsync/redsync/v4"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type Lock struct {
	mutex *redsync.Mutex
}

// DistributedCache is an instance of a distributed cache
type DistributedCache interface {
	Get(ctx context.Context, key string, out interface{}) error
	Save(key string, value any, expiration time.Duration) error

	Delete(key string) error

	// AcquireLock tries to acquire a lock with given key
	AcquireLock(ctx context.Context, key string) (*Lock, error)

	// ReleaseLock releases lock
	ReleaseLock(lock *Lock) error
}

type Mock struct{}

func (m *Mock) Get(ctx context.Context, key string, out interface{}) error { return nil }
func (m *Mock) Save(key string, value any, expiration time.Duration) error { return nil }

func (m *Mock) Delete(key string) error                                    { return nil }
func (m *Mock) AcquireLock(ctx context.Context, key string) (*Lock, error) { return &Lock{}, nil }

func (m *Mock) ReleaseLock(lock *Lock) error { return nil }

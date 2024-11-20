package middlewares

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

func mustConnectRedis() *redis.Client {
	redisHost := os.Getenv("REDIS_HOST")
	client := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	if _, err := client.Ping(context.Background()).Result(); err != nil {
		panic(fmt.Errorf("failed to connect to Redis: %w", err))
	}

	return client
}

// TestDistributedRateLimiter_Middleware is an integration test that tests
// how DistributedRateLimiter would behave in a live environment by using
// real dependencies rather than mock the dependencies.
// It is therefore recommended to run this test in a containerised environment
// where the test environment can be setup comfortably
func TestDistributedRateLimiter_Middleware(t *testing.T) {
	redisClient := mustConnectRedis()
	defer redisClient.Close()

	drl := NewDistributedRateLimiter(redisClient, 5, time.Second*10)

	handler := drl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const numOfClients = 20
	const requestPerClient = 3

	var wg sync.WaitGroup
	wg.Add(numOfClients)

	results := make(chan int, numOfClients*requestPerClient)

	for i := 0; i < numOfClients; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < requestPerClient; j++ {

				// Random sleep to simulate network delay
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))

				req := httptest.NewRequest(http.MethodGet, "/", nil)
				w := httptest.NewRecorder()

				handler.ServeHTTP(w, req)

				results <- w.Result().StatusCode
			}
		}()
	}

	wg.Wait()
	close(results)

	// Analyze results
	var withinLimit, exceededLimit int
	for result := range results {
		if result == http.StatusOK {
			withinLimit++
		} else if result == http.StatusTooManyRequests {
			exceededLimit++
		}
	}

	if exceededLimit == 0 {
		t.Errorf("Expected some requests to exceed the rate limit")
	}
	if withinLimit > 5 {
		t.Errorf("Expected only 5 requests to pass within the rate limit window, but got %v", withinLimit)
	}
}

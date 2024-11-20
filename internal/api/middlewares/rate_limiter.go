package middlewares

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/http"
	"time"
)

const serviceName = "payment-gateways"

// DistributedRateLimiter implements the fixed-window counter rate limiting algorithm as highlighted below:
//
//   - Each unique client (identified by IP address) is given a counter in the distributed storage.
//   - Counter increments with every request.
//   - If counter exceeds pre-defined limit within the fixed time window (e.g., 1 minute),
//     further requests are denied until the window resets.
//   - Counter is reset after each window expires by setting an expiration time on the key.
type DistributedRateLimiter struct {
	client *redis.Client

	// Maximum number of requests allowed within window
	requestLimit int

	// Time window for rate limiting
	window time.Duration
}

func NewDistributedRateLimiter(client *redis.Client, requestLimit int, window time.Duration) *DistributedRateLimiter {
	return &DistributedRateLimiter{
		client:       client,
		requestLimit: requestLimit,
		window:       window,
	}
}

func (rl *DistributedRateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		allowed, err := rl.allow(r.Context(), userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request is within the allowed limit for a given IP
func (rl *DistributedRateLimiter) allow(ctx context.Context, key string) (bool, error) {

	key = rl.constructRateLimitKey(key)

	// Increment the count for the current window
	count, err := rl.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set an expiration on the key if this is the first request
	if count == 1 {
		if err = rl.client.Expire(ctx, key, rl.window).Err(); err != nil {
			return false, err
		}
	}

	// Check if the count exceeds the limit
	return count <= int64(rl.requestLimit), nil
}

func (rl *DistributedRateLimiter) constructRateLimitKey(key string) string {
	return fmt.Sprintf("ratelimit_%s:%s", serviceName, key)
}

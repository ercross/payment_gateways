package services

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker configuration for kafka publisher doing only 1 request at a time each five seconds
var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
	Name:        "KafkaPublisher",
	MaxRequests: 1,
	Interval:    5 * time.Second,
	Timeout:     3 * time.Second,
})

func PublishWithCircuitBreaker(operation func() error) error {
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, operation()
	})
	return err
}

func RetryOperation(operation func() error, maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(2^i) * time.Second)
	}
	return fmt.Errorf("operation failed after %d attempts", maxRetries)
}

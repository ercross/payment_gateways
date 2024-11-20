package kafka

import (
	"context"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
)

type EventPublisher interface {
	PublishTransaction(ctx context.Context, transactionID int, message []byte, dataFormat dto.DataFormat) error
	Close() error
}

type Mock struct{}

func (m *Mock) PublishTransaction(ctx context.Context, transactionID int, message []byte, dataFormat dto.DataFormat) error {
	return nil
}
func (m *Mock) Close() error { return nil }

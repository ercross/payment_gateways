package kafka

import (
	"context"
	"fmt"
	"github.com/ercross/payment_gateways/internal/api/v1/dto"
	"github.com/ercross/payment_gateways/internal/services"
	"time"

	"github.com/segmentio/kafka-go"
)

type Kafka struct {
	writer *kafka.Writer
}

func NewProducer(brokerUrl string) *Kafka {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(brokerUrl),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           10 * time.Millisecond,
	}
	return &Kafka{
		writer: writer,
	}
}

// GetTopic returns the appropriate Kafka topic based on the data format.
func GetTopic(dataFormat dto.DataFormat) (string, error) {
	switch dataFormat {
	case dto.DataFormatJSON:
		return "transactions.json", nil
	case dto.DataFormatXML:
		return "transactions.soap", nil
	default:
		return "", fmt.Errorf("unsupported data format: %s", dataFormat)
	}
}

// PublishTransaction publishes to transaction topic
func (p *Kafka) PublishTransaction(ctx context.Context, transactionID int, message []byte, dataFormat dto.DataFormat) error {

	topic, err := GetTopic(dataFormat)
	if err != nil {
		return err
	}

	kafkaMessage := kafka.Message{
		Key:   []byte(fmt.Sprint(transactionID)),
		Value: message,
		Topic: topic,
	}

	err = services.PublishWithCircuitBreaker(func() error {
		return p.writer.WriteMessages(ctx, kafkaMessage)
	})
	if err != nil {
		return fmt.Errorf("error publishing transaction %d: %w", transactionID, err)
	}

	return nil
}

// Close the writer when the system shut down
func (p *Kafka) Close() error {
	return p.writer.Close()
}

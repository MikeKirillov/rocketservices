package order

import (
	"context"

	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

// KafkaProducer определяет контракт для отправки сообщений в Kafka.
type KafkaProducer interface {
	Send(ctx context.Context, msg *kafka.Message) error
}

package order

import (
	"context"

	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

type KafkaProducer interface {
	Send(ctx context.Context, msg *kafka.Message) error
}

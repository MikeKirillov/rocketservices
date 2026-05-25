package assembly

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

type OrderService interface {
	Commit(ctx context.Context, id uuid.UUID) error
}

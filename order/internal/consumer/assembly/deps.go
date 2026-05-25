package assembly

import (
	"context"

	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

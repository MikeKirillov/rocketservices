package assembly

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/platform/pkg/kafka"
)

type Consumer interface {
	Consume(ctx context.Context, handler kafka.MessageHandler) error
}

type OrderRepository interface {
	GetForUpdate(ctx context.Context, id uuid.UUID) (model.Order, error)
	Update(ctx context.Context, order model.Order) error
}

type InventoryServiceClient interface {
	CommitParts(ctx context.Context, order model.Order) error
}

type TxManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

type OrderRepository interface {
	Get(ctx context.Context, id uuid.UUID) (model.Order, error)
	GetForUpdate(ctx context.Context, id uuid.UUID) (model.Order, error)
	Create(ctx context.Context, order model.Order) error
	Update(ctx context.Context, order model.Order) error
}

type InventoryServiceClient interface {
	ListParts(ctx context.Context, partsUuids []uuid.UUID) ([]model.OrderItem, error)
	ValidateCompatibility(ctx context.Context, items []model.OrderItem) error
	ReserveParts(ctx context.Context, partsUuids []uuid.UUID) error
	ReleaseParts(ctx context.Context, order model.Order) error
	CommitParts(ctx context.Context, order model.Order) error
}

type PaymentServiceClient interface {
	PayOrder(ctx context.Context, orderRq input.PayOrderRq) (uuid.UUID, error)
}

type AuthServiceClient interface {
	Whoami(ctx context.Context, sessionUUID uuid.UUID) (uuid.UUID, uuid.UUID, error)
}

type OrderProducerService interface {
	ProduceOrderPaid(ctx context.Context, order model.Order) error
}

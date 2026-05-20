package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

type OrderRepository interface {
	Get(ctx context.Context, id uuid.UUID) (model.Order, error)
	Create(ctx context.Context, order model.Order) error
	Update(ctx context.Context, order model.Order) error
}

type InventoryServiceClient interface {
	ListParts(ctx context.Context, partsUuids []uuid.UUID) ([]model.OrderItem, error)
}

type PaymentServiceClient interface {
	PayOrder(ctx context.Context, orderRq model.PayOrderRq) (uuid.UUID, error)
}

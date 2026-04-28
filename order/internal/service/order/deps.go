package order

import (
	"context"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

type OrderRepository interface {
	Get(ctx context.Context, id string) (model.Order, error)
	Create(ctx context.Context, order record.Order) error
	Update(ctx context.Context, order record.Order) error
}

type InventoryServiceClient interface {
	ListParts(ctx context.Context, partsUuids []string) ([]model.Part, error)
}

type PaymentServiceClient interface {
	PayOrder(ctx context.Context, orderRq model.PayOrderRq) (string, error)
}

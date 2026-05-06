package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
)

type OrderService interface {
	Get(ctx context.Context, id uuid.UUID) (model.Order, error)
	Create(ctx context.Context, orderRq model.CreateOrderRq) (model.CreateOrderRs, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	Pay(ctx context.Context, orderRq model.PayOrderRq) (uuid.UUID, error)
}

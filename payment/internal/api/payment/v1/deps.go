package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/payment/internal/model"
)

type OrderService interface {
	Pay(ctx context.Context, order model.Order) (string, error)
}

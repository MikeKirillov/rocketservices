package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/payment/internal/model"
)

type OrderService interface {
	Pay(ctx context.Context, order model.Payment) (uuid.UUID, error)
}

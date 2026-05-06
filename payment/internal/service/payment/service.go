package payment

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/payment/internal/model"
)

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) Pay(_ context.Context, order model.Payment) (uuid.UUID, error) {
	// Сгенерировать transaction_uuid (UUID v4)
	newUUID := uuid.New()

	slog.Info("оплата выполнена", "order_uuid", order.Uuid, "transaction_uuid", newUUID, "payment_method", order.PaymentMethod)

	return newUUID, nil
}

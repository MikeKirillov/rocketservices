package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/payment/internal/errors"
	"github.com/MikeKirillov/rocketservices/payment/internal/model"
)

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) Pay(_ context.Context, order model.Order) (string, error) {
	// Проверить, что order_uuid не пустой → INVALID_ARGUMENT
	if order.Uuid == "" {
		return "", errs.ErrOrderOmittedUUID
	}

	// Проверить, что payment_method != UNSPECIFIED → INVALID_ARGUMENT
	if !order.PaymentMethod.IsValid() {
		return "", fmt.Errorf("способ оплаты %s: %w", order.PaymentMethod, errs.ErrOrderInvalidPayMethod)
	}

	// Проверить формат UUID → INVALID_ARGUMENT
	if _, err := uuid.Parse(order.Uuid); err != nil {
		return "", fmt.Errorf("заказ %s: %w", order.Uuid, errs.ErrOrderInvalidUUID)
	}

	// Сгенерировать transaction_uuid (UUID v4)
	newUUID := uuid.NewString()

	// Вывести в лог: "оплата прошла успешно, order_uuid: X, transaction_uuid: Y"
	slog.Info("оплата прошла успешно", "order_uuid", order.Uuid, "transaction_uuid", newUUID)

	return newUUID, nil
}

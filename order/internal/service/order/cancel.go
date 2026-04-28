package order

import (
	"context"
	"fmt"
	"log/slog"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
)

func (s *service) Cancel(ctx context.Context, id string) error {
	// 1. Найти заказ в store
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		slog.Error("заказ не найден", "uuid", id)
		return errs.ErrOrderNotFound
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != model.OrderStatusPendingPayment {
		slog.Error("невозможно отменить заказ", "текущий статус", order.Status)
		return errs.ErrOrderCancelling
	}

	// 3. Обновить статус на CANCELLED
	err = s.orderRepository.Update(ctx, converter.OrderToRepoUpdate(order, model.OrderStatusCancelled, nil, nil))
	if err != nil {
		slog.Error("заказ не обновлен", "uuid", id)
		return fmt.Errorf("обновить заявку: %w", err)
	}
	return nil
}

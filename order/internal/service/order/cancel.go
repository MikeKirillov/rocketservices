package order

import (
	"context"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Cancel(ctx context.Context, id string) error {
	// 1. Найти заказ в store
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		return sharedErrs.NewNotFound(err)
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != model.OrderStatusPendingPayment {
		return sharedErrs.NewConflict(errs.ErrOrderCancelling)
	}

	// 3. Обновить статус на CANCELLED
	order.SetStatus(model.OrderStatusCancelled)

	orderRepo := converter.OrderToRepoUpdate(order)
	if err = s.orderRepository.Update(ctx, orderRepo); err != nil {
		return sharedErrs.NewNotFound(err)
	}
	return nil
}

package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	// Найти заказ в store
	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return sharedErrs.NewNotFound(err)
		}
		return sharedErrs.NewInternal(err)
	}

	// Отменить резерв деталей на складе
	err = s.inventoryClient.ReleaseParts(ctx, order)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrGrpcNotFound):
			return sharedErrs.NewNotFound(err)
		case errors.Is(err, errs.ErrGrpcInvalidArgument):
			return sharedErrs.NewInvalidArgument(err)
		default:
			return sharedErrs.NewInternal(err)
		}
	}

	// Обновить статус на CANELLED
	err = order.Cancel()
	if err != nil {
		return sharedErrs.NewConflict(err)
	}
	if err = s.orderRepository.Update(ctx, order); err != nil {
		if errors.Is(err, errs.ErrOrderConflict) {
			return sharedErrs.NewConflict(err)
		}
		return sharedErrs.NewInternal(err)
	}
	return nil
}

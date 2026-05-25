package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Commit(ctx context.Context, id uuid.UUID) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.GetForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				return sharedErrs.NewNotFound(err)
			}
			return sharedErrs.NewInternal(err)
		}

		// Получить заказ из БД и проверить статус
		// Если заказ уже ASSEMBLED — пропустить (идемпотентность, at-least-once семантика)
		err = order.CanCommit()
		if err != nil {
			return sharedErrs.NewConflict(err)
		}

		// Вызвать CommitParts через gRPC-клиент InventoryService
		err = s.inventoryClient.CommitParts(ctx, order)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrGrpcNotFound):
				return sharedErrs.NewNotFound(err)
			case errors.Is(err, errs.ErrGrpcInvalidArgument):
				return sharedErrs.NewInvalidArgument(err)
			case errors.Is(err, errs.ErrGrpcCancelledOperation):
				return sharedErrs.NewFailedPrecondition(err)
			default:
				return sharedErrs.NewInternal(err)
			}
		}

		// Обновить статус заказа на ASSEMBLED
		order.Commit()
		if err = s.orderRepository.Update(ctx, order); err != nil {
			if errors.Is(err, errs.ErrOrderConflict) {
				return sharedErrs.NewConflict(err)
			}
			return sharedErrs.NewInternal(err)
		}

		return nil
	})
	return err
}

package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Cancel(ctx context.Context, id uuid.UUID) error {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Cancel", trace.WithAttributes(
		attribute.String("order.uuid", id.String()),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "order.Cancel.Ok")
		}
		span.End()
	}()

	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.GetForUpdate(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				resultErr = sharedErrs.NewNotFound(err)
				return resultErr
			}
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}

		// Отменить резерв деталей на складе
		err = s.inventoryClient.ReleaseParts(ctx, order)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrGrpcNotFound):
				{
					resultErr = sharedErrs.NewNotFound(err)
					return resultErr
				}
			case errors.Is(err, errs.ErrGrpcInvalidArgument):
				{
					resultErr = sharedErrs.NewInvalidArgument(err)
					return resultErr
				}
			case errors.Is(err, errs.ErrGrpcCancelledOperation):
				{
					resultErr = sharedErrs.NewFailedPrecondition(err)
					return resultErr
				}
			default:
				{
					resultErr = sharedErrs.NewInternal(err)
					return resultErr
				}
			}
		}

		// Обновить статус на CANELLED
		err = order.Cancel()
		if err != nil {
			resultErr = sharedErrs.NewConflict(err)
			return resultErr
		}
		if err = s.orderRepository.Update(ctx, order); err != nil {
			if errors.Is(err, errs.ErrOrderConflict) {
				resultErr = sharedErrs.NewConflict(err)
				return resultErr
			}
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}
		return nil
	})
}

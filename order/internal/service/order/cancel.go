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
	defer span.End()

	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.GetForUpdate(ctx, id)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			if errors.Is(err, errs.ErrOrderNotFound) {
				return sharedErrs.NewNotFound(err)
			}
			return sharedErrs.NewInternal(err)
		}

		// Отменить резерв деталей на складе
		err = s.inventoryClient.ReleaseParts(ctx, order)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

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

		// Обновить статус на CANELLED
		err = order.Cancel()
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return sharedErrs.NewConflict(err)
		}
		if err = s.orderRepository.Update(ctx, order); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			if errors.Is(err, errs.ErrOrderConflict) {
				return sharedErrs.NewConflict(err)
			}
			return sharedErrs.NewInternal(err)
		}
		return nil
	})
}

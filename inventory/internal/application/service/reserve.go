package part

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Reserve(ctx context.Context, ids []uuid.UUID) error {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Reserve", trace.WithAttributes(
		attribute.StringSlice("item.uuids", uuidsToStringSlice(ids)),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "inventory.Reserve.Ok")
		}
		span.End()
	}()

	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// Загрузите детали по uuids через репозиторий.
		parts, err := s.partRepository.ListByUUIDsForUpdate(ctx, ids)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrPartNotFound):
				{
					resultErr = sharedErrs.NewNotFound(err)
					return resultErr
				}
			case errors.Is(err, errs.ErrWrongPartType):
				{
					resultErr = sharedErrs.NewConflict(err)
					return resultErr
				}
			default:
				{
					resultErr = sharedErrs.NewInternal(err)
					return resultErr
				}
			}
		}
		// Вызовите соответствующий метод entity (Reserve() / Release()) для каждой детали —
		// entity сама вернёт ошибку при нарушении инварианта.
		for i := range parts {
			if err = parts[i].Reserve(); err != nil {
				resultErr = sharedErrs.NewResourceExhausted(err)
				return resultErr
			}
		}
		// Сохраните обновлённое состояние через UpdateReservedBatch — один SQL-запрос для всех деталей.
		if err = s.partRepository.UpdateReservedBatch(ctx, parts); err != nil {
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}
		return nil
	})
}

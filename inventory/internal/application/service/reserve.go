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
	defer span.End()

	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// Загрузите детали по uuids через репозиторий.
		parts, err := s.partRepository.ListByUUIDsForUpdate(ctx, ids)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			switch {
			case errors.Is(err, errs.ErrPartNotFound):
				return sharedErrs.NewNotFound(err)
			case errors.Is(err, errs.ErrWrongPartType):
				return sharedErrs.NewConflict(err)
			default:
				return sharedErrs.NewInternal(err)
			}
		}
		// Вызовите соответствующий метод entity (Reserve() / Release()) для каждой детали —
		// entity сама вернёт ошибку при нарушении инварианта.
		for i := range parts {
			if err = parts[i].Reserve(); err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				return sharedErrs.NewResourceExhausted(err)
			}
		}
		// Сохраните обновлённое состояние через UpdateReservedBatch — один SQL-запрос для всех деталей.
		if err = s.partRepository.UpdateReservedBatch(ctx, parts); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())

			return sharedErrs.NewInternal(err)
		}
		return nil
	})
}

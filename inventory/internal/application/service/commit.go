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

func (s *service) Commit(ctx context.Context, ids []uuid.UUID) error {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Commit", trace.WithAttributes(
		attribute.StringSlice("item.uuids", uuidsToStringSlice(ids)),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "inventory.Commit.Ok")
		}
		span.End()
	}()

	return s.txManager.Do(ctx, func(ctx context.Context) error {
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

		for i := range parts {
			if err = parts[i].Commit(); err != nil {
				resultErr = sharedErrs.NewFailedPrecondition(err)
				return resultErr
			}
		}

		if err = s.partRepository.UpdateCommitBatch(ctx, parts); err != nil {
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}
		return nil
	})
}

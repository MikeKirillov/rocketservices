package part

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/inventory/internal/domain/model"
	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (model.Part, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("inventory-service").Start(ctx, "inventory.Get", trace.WithAttributes(
		attribute.String("item.uuid", id.String()),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "inventory.Get.Ok")
		}
		span.End()
	}()

	// Проверить, что uuid не пустой → INVALID_ARGUMENT
	if id == uuid.Nil {
		return model.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartOmittedUUID)
	}
	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrPartNotFound) {
			resultErr = sharedErrs.NewNotFound(err)
			return model.Part{}, resultErr
		}
		resultErr = sharedErrs.NewInternal(err)
		return model.Part{}, resultErr
	}

	span.SetAttributes(
		attribute.String("item.uuid", part.UUID().String()),
		attribute.String("item.name", part.Name()),
		attribute.String("item.type", string(part.PartType())),
		attribute.Int("item.quantity", part.StockQuantity()),
		attribute.Int64("item.price", part.Price()),
	)

	return part, nil
}

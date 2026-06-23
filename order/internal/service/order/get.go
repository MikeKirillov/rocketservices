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
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (model.Order, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Get", trace.WithAttributes(
		attribute.String("order.uuid", id.String()),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "order.Get.Ok")
		}
		span.End()
	}()

	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			resultErr = sharedErrs.NewNotFound(err)
			return model.Order{}, resultErr
		}
		resultErr = sharedErrs.NewInternal(err)
		return model.Order{}, resultErr
	}

	span.SetAttributes(
		attribute.String("order.uuid", order.OrderUUID.String()),
		attribute.Int("order.items_count", len(order.Items)),
	)

	return order, nil
}

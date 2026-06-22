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
	defer span.End()

	order, err := s.orderRepository.Get(ctx, id)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, errs.ErrOrderNotFound) {
			return model.Order{}, sharedErrs.NewNotFound(err)
		}
		return model.Order{}, sharedErrs.NewInternal(err)
	}

	span.SetAttributes(
		attribute.String("order.uuid", order.OrderUUID.String()),
		attribute.Int("order.items_count", len(order.Items)),
	)

	return order, nil
}

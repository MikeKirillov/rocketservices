package payment

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/payment/internal/model"
)

type service struct{}

func NewService() *service {
	return &service{}
}

func (s *service) Pay(ctx context.Context, order model.Payment) (uuid.UUID, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("payment-service").Start(ctx, "payment.Pay", trace.WithAttributes(
		attribute.String("order.uuid", order.Uuid.String()),
		attribute.String("payment.type", string(order.PaymentMethod)),
	))
	defer func() {
		span.SetStatus(codes.Ok, "payment.Pay.Ok")
		span.End()
	}()

	// Сгенерировать transaction_uuid (UUID v4)
	newUUID := uuid.New()

	slog.Info("оплата выполнена",
		slog.String("order_uuid", order.Uuid.String()),
		slog.String("transaction_uuid", newUUID.String()),
		slog.String("payment_method", string(order.PaymentMethod)),
	)

	span.SetAttributes(attribute.String("transaction.uuid", newUUID.String()))

	return newUUID, nil
}

package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Pay(ctx context.Context, orderRq input.PayOrderRq) (uuid.UUID, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("order-service").Start(ctx, "order.Pay", trace.WithAttributes(
		attribute.String("order.uuid", orderRq.OrderUUID.String()),
		attribute.String("order.payment_method", string(orderRq.PaymentMethod)),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "order.Pay.Ok")
		}
		span.End()
	}()

	var transactionResult uuid.UUID
	var totalPrice int64

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.GetForUpdate(ctx, orderRq.OrderUUID)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				resultErr = sharedErrs.NewNotFound(err)
				return resultErr
			}
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}

		// Проверить статус == PENDING_PAYMENT
		err = order.CanModify()
		if err != nil {
			resultErr = sharedErrs.NewConflict(err)
			return resultErr
		}

		// Вызвать h.paymentClient.PayOrder для обработки платежа
		transactionUUIDRs, err := s.paymentClient.PayOrder(ctx, orderRq)
		if err != nil {
			resultErr = err // клиент уже завернул в BusinessError у себя
			return resultErr
		}

		// Обновить статус на PAID и сохранить transaction_uuid
		// и payment_method
		order.Pay(transactionUUIDRs, &orderRq.PaymentMethod)
		if err = s.orderRepository.Update(ctx, order); err != nil {
			if errors.Is(err, errs.ErrOrderConflict) {
				resultErr = sharedErrs.NewConflict(err)
				return resultErr
			}
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}

		// Отправка события OrderPaid в Kafka
		err = s.orderProducerService.ProduceOrderPaid(ctx, order)
		if err != nil {
			resultErr = sharedErrs.NewInternal(err)
			return resultErr
		}

		// Вернуть transaction_uuid
		transactionResult = transactionUUIDRs
		// Вернуть totalPrice для передачи метрик
		totalPrice = order.TotalPrice
		return nil
	})
	if err != nil {
		resultErr = err
		return uuid.Nil, resultErr
	}

	// Фиксируем метрики для отображения в Grafana
	ordersPaid.Add(ctx, 1)
	ordersRevenue.Add(ctx, totalPrice)

	span.SetAttributes(attribute.String("order.transaction_uuid", transactionResult.String()))

	return transactionResult, nil
}

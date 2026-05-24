package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Pay(ctx context.Context, orderRq input.PayOrderRq) (uuid.UUID, error) {
	var transactionResult uuid.UUID

	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		// Найти заказ в store
		order, err := s.orderRepository.Get(ctx, orderRq.OrderUUID)
		if err != nil {
			if errors.Is(err, errs.ErrOrderNotFound) {
				return sharedErrs.NewNotFound(err)
			}
			return sharedErrs.NewInternal(err)
		}

		// Проверить статус == PENDING_PAYMENT
		err = order.CanModify()
		if err != nil {
			return sharedErrs.NewConflict(err)
		}

		// Вызвать h.paymentClient.PayOrder для обработки платежа
		transactionUUIDRs, err := s.paymentClient.PayOrder(ctx, orderRq)
		if err != nil {
			return err // клиент уже завернул в BusinessError у себя
		}

		// Обновить статус на PAID и сохранить transaction_uuid
		// и payment_method
		order.Pay(transactionUUIDRs, &orderRq.PaymentMethod)
		if err = s.orderRepository.Update(ctx, order); err != nil {
			if errors.Is(err, errs.ErrOrderConflict) {
				return sharedErrs.NewConflict(err)
			}
			return sharedErrs.NewInternal(err)
		}

		// Отправка события OrderPaid в Kafka
		err = s.orderProducerService.ProduceOrderPaid(ctx, order)
		if err != nil {
			return sharedErrs.NewInternal(err)
		}

		// Вернуть transaction_uuid
		transactionResult = transactionUUIDRs
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return transactionResult, nil
}

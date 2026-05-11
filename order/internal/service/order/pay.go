package order

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Pay(ctx context.Context, orderRq model.PayOrderRq) (uuid.UUID, error) {
	// 1. Найти заказ в store
	order, err := s.orderRepository.Get(ctx, orderRq.OrderUUID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return uuid.Nil, sharedErrs.NewNotFound(err)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	// 2. Проверить статус == PENDING_PAYMENT
	err = order.CanModify()
	if err != nil {
		return uuid.Nil, sharedErrs.NewConflict(err)
	}

	// 3. Вызвать h.paymentClient.PayOrder для обработки платежа
	transactionUUIDRs, err := s.paymentClient.PayOrder(ctx, orderRq)
	if err != nil {
		return uuid.Nil, err // клиент уже завернул в BusinessError у себя
	}

	// 4. Обновить статус на PAID и сохранить transaction_uuid
	// и payment_method
	order.Pay(transactionUUIDRs, &orderRq.PaymentMethod)
	if err = s.orderRepository.Update(ctx, order); err != nil {
		if errors.Is(err, errs.ErrOrderConflict) {
			return uuid.Nil, sharedErrs.NewConflict(err)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	// 5. Вернуть transaction_uuid
	return transactionUUIDRs, nil
}

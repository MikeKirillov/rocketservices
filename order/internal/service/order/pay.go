package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Pay(ctx context.Context, orderRq model.PayOrderRq) (uuid.UUID, error) {
	// 1. Найти заказ в store
	order, err := s.orderRepository.Get(ctx, orderRq.OrderUUID)
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return uuid.Nil, sharedErrs.NewNotFound(err)
		}
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != model.OrderStatusPendingPayment {
		return uuid.Nil, sharedErrs.NewConflict(errs.ErrOrderAlreadyPaid)
	}

	// 3. Вызвать h.paymentClient.PayOrder для обработки платежа
	transactionUUIDRs, err := s.paymentClient.PayOrder(ctx, orderRq)
	if err != nil {
		return uuid.Nil, err // клиент уже завернул в BusinessError у себя
	}

	// 4. Обновить статус на PAID и сохранить transaction_uuid
	// и payment_method
	transactionUUID, err := uuid.Parse(transactionUUIDRs)
	if err != nil {
		return uuid.Nil, sharedErrs.NewInvalidArgument(fmt.Errorf("транзакция %s: %w", transactionUUID, errs.ErrInvalidUUID))
	}

	order.SetPaymentMethod(&orderRq.PaymentMethod)
	order.SetTransaction(transactionUUIDRs)
	order.SetStatus(model.OrderStatusPaid)

	orderRepo := converter.OrderToRepoUpdate(order)
	if err = s.orderRepository.Update(ctx, orderRepo); err != nil {
		return uuid.Nil, sharedErrs.NewNotFound(err)
	}

	// 5. Вернуть transaction_uuid
	return transactionUUID, nil
}

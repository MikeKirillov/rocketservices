package order

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/converter"
)

func (s *service) Pay(ctx context.Context, orderRq model.PayOrderRq) (uuid.UUID, error) {
	// 1. Найти заказ в store
	order, err := s.orderRepository.Get(ctx, orderRq.OrderUUID)
	if err != nil {
		slog.Error("заказ не найден", "uuid", orderRq.OrderUUID)
		return uuid.Nil, errs.ErrOrderNotFound
	}

	// 2. Проверить статус == PENDING_PAYMENT
	if order.Status != model.OrderStatusPendingPayment {
		slog.Error("невозможно оплатить заказ", "текущий статус", order.Status)
		return uuid.Nil, errs.ErrOrderAlreadyPaid
	}

	// 3. Вызвать h.paymentClient.PayOrder для обработки платежа
	transactionUUIDRs, err := s.paymentClient.PayOrder(ctx, orderRq)
	if err != nil {
		return uuid.Nil, err
	}

	// 4. Обновить статус на PAID и сохранить transaction_uuid
	// и payment_method
	transactionUUID, err := uuid.Parse(transactionUUIDRs)
	if err != nil {
		slog.Error("неверный формат", "uuid", transactionUUID)
		return uuid.Nil, fmt.Errorf("транзакция %s: %w", transactionUUID, errs.ErrInvalidUUID)
	}
	err = s.orderRepository.Update(ctx, converter.OrderToRepoUpdate(order, model.OrderStatusPaid, &transactionUUIDRs, &orderRq.PaymentMethod))
	if err != nil {
		slog.Error("заказ не обновлен", "uuid", order.OrderUUID)
		return uuid.Nil, fmt.Errorf("обновить заявку: %w", err)
	}

	// 5. Вернуть transaction_uuid
	return transactionUUID, nil
}

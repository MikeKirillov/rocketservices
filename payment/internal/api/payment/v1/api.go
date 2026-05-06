package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/payment/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/payment/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

type api struct {
	paymentv1.UnimplementedPaymentServiceServer

	orderService OrderService
}

func NewAPI(orderService OrderService) *api {
	return &api{orderService: orderService}
}

// PayOrder обрабатывает оплату заказа.
func (a *api) PayOrder(ctx context.Context, req *paymentv1.PayOrderRequest) (*paymentv1.PayOrderResponse, error) {
	// Проверить, что order_uuid не пустой → INVALID_ARGUMENT
	if req.GetOrderUuid() == "" {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrPaymentOmittedUUID)
	}
	// Проверить формат UUID → INVALID_ARGUMENT
	// Проверить, что payment_method != UNSPECIFIED → INVALID_ARGUMENT
	payment, err := converter.PaymentToModel(req)
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(err)
	}
	transaction, err := a.orderService.Pay(ctx, payment)
	if err != nil {
		return nil, err
	}

	// Вернуть transaction_uuid
	return &paymentv1.PayOrderResponse{
		TransactionUuid: transaction.String(),
	}, nil
}

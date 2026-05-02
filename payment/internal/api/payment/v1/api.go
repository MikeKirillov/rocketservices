package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/payment/internal/converter"
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
	order := converter.OrderToModel(req)
	transaction, err := a.orderService.Pay(ctx, order)
	if err != nil {
		return nil, err
	}

	// Вернуть transaction_uuid
	return &paymentv1.PayOrderResponse{
		TransactionUuid: transaction,
	}, nil
}

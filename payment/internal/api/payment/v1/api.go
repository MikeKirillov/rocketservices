package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/MikeKirillov/rocketservices/payment/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/payment/internal/errors"
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
	order := converter.OrderToModel(req.GetOrderUuid(), req.GetPaymentMethod())
	transaction, err := a.orderService.Pay(ctx, order)
	if err != nil {
		if errors.Is(err, errs.ErrOrderOmittedUUID) ||
			errors.Is(err, errs.ErrOrderInvalidUUID) ||
			errors.Is(err, errs.ErrOrderInvalidPayMethod) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "внутрення ошибка: %v", err)
	}

	// Вернуть transaction_uuid
	return &paymentv1.PayOrderResponse{
		TransactionUuid: transaction,
	}, nil
}

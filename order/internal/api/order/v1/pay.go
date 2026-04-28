package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// PayOrder реализует операцию payOrder.
// POST /api/v1/orders/{order_uuid}/pay.
func (o *orderHandler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	payOrderRq, err := converter.PayOrderRqToModel(params.OrderUUID, req.PaymentMethod)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidPaymentMethod) {
			return &orderv1.PayOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil
		}
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}

	transactionUUID, err := o.orderService.Pay(ctx, payOrderRq)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrOrderNotFound):
			return &orderv1.PayOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil

		case errors.Is(err, errs.ErrOrderAlreadyPaid) || errors.Is(err, errs.ErrInvalidUUID):
			return &orderv1.PayOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil

		default:
			return &orderv1.PayOrderInternalServerError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}
	}

	// 5. Вернуть transaction_uuid
	return &orderv1.PayOrderResponse{TransactionUUID: transactionUUID}, nil
}

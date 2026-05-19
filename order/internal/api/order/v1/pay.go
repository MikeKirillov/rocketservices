package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/order/internal/api/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// PayOrder реализует операцию payOrder.
// POST /api/v1/orders/{order_uuid}/pay.
func (o *orderHandler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	payOrderRq, err := converter.PayOrderRqToModel(params.OrderUUID, req.PaymentMethod)
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.PayOrderRes](
			err,
			sharedErrs.WithBadRequest[orderv1.PayOrderBadRequest](),
			sharedErrs.WithInternal[orderv1.PayOrderInternalServerError](),
		)
	}

	transactionUUID, err := o.orderService.Pay(ctx, payOrderRq)
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.PayOrderRes](
			err,
			sharedErrs.WithNotFound[orderv1.PayOrderNotFound](),
			sharedErrs.WithConflict[orderv1.PayOrderConflict](),
			sharedErrs.WithInternal[orderv1.PayOrderInternalServerError](),
		)
	}

	// 5. Вернуть transaction_uuid
	return &orderv1.PayOrderResponse{TransactionUUID: transactionUUID}, nil
}

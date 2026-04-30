package v1

import (
	"context"

	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// CancelOrder реализует операцию cancelOrder.
// POST /api/v1/orders/{order_uuid}/cancel.
func (o *orderHandler) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	err := o.orderService.Cancel(ctx, params.OrderUUID.String())
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.CancelOrderRes](
			err,
			sharedErrs.WithNotFound[orderv1.CancelOrderNotFound](),
			sharedErrs.WithConflict[orderv1.CancelOrderConflict](),
		)
	}

	// Вернуть success
	return &orderv1.CancelOrderResponse{}, nil
}

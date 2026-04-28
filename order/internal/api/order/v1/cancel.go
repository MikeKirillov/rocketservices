package v1

import (
	"context"
	"errors"
	"net/http"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// CancelOrder реализует операцию cancelOrder.
// POST /api/v1/orders/{order_uuid}/cancel.
func (o *orderHandler) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	err := o.orderService.Cancel(ctx, params.OrderUUID.String())
	if errors.Is(err, errs.ErrOrderNotFound) {
		return &orderv1.CancelOrderNotFound{
			Code:    http.StatusNotFound,
			Message: err.Error(),
		}, nil
	}
	if errors.Is(err, errs.ErrOrderCancelling) {
		return &orderv1.CancelOrderConflict{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}, nil
	}

	// Вернуть success
	return &orderv1.CancelOrderResponse{}, nil
}

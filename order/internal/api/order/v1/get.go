package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (o *orderHandler) GetOrder(ctx context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	order, err := o.orderService.Get(ctx, params.OrderUUID.String())
	if err != nil {
		if errors.Is(err, errs.ErrOrderNotFound) {
			return &orderv1.GetOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil
		}
		return &orderv1.GetOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}, nil
	}

	orderRs, err := converter.OrderToDTO(order)
	if err != nil {
		return &orderv1.GetOrderInternalServerError{
			Code:    http.StatusConflict,
			Message: err.Error(),
		}, nil
	}

	return orderRs, nil
}

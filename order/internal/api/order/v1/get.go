package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (o *orderHandler) GetOrder(ctx context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	order, err := o.orderService.Get(ctx, params.OrderUUID.String())
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.GetOrderRes](
			err,
			sharedErrs.WithNotFound[orderv1.GetOrderNotFound](),
			sharedErrs.WithInternal[orderv1.GetOrderInternalServerError](),
		)
	}

	orderRs, err := converter.OrderToDTO(order)
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.GetOrderRes](
			err,
			sharedErrs.WithInternal[orderv1.GetOrderInternalServerError](),
		)
	}

	return orderRs, nil
}

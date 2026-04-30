package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// CreateOrder реализует операцию createOrder.
// POST /api/v1/orders.
func (o *orderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	orderRq := converter.OrderCreateRqToModel(req)

	order, err := o.orderService.Create(ctx, orderRq)
	if err != nil {
		return sharedErrs.MapHTTPError[orderv1.CreateOrderRes](
			err,
			sharedErrs.WithBadRequest[orderv1.CreateOrderBadRequest](),
			sharedErrs.WithNotFound[orderv1.CreateOrderNotFound](),
			sharedErrs.WithConflict[orderv1.CreateOrderConflict](),
			sharedErrs.WithInternal[orderv1.CreateOrderInternalServerError](),
		)
	}

	// 8. Вернуть order_uuid и total_price
	return converter.OrderCreateRsToDTO(order), nil
}

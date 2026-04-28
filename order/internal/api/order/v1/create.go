package v1

import (
	"context"
	"errors"
	"net/http"

	"github.com/MikeKirillov/rocketservices/order/internal/converter"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// CreateOrder реализует операцию createOrder.
// POST /api/v1/orders.
func (o *orderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	orderRq := converter.OrderCreateRqToModel(req)

	order, err := o.orderService.Create(ctx, orderRq)
	if err != nil {
		switch {
		case errors.Is(err, errs.ErrOmittedRequiredField) ||
			errors.Is(err, errs.ErrInvalidUUID) ||
			errors.Is(err, errs.ErrGrpcInvalidArgument):
			return &orderv1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: err.Error(),
			}, nil

		case errors.Is(err, errs.ErrPartNotFound) || errors.Is(err, errs.ErrInventoryNotFound):
			return &orderv1.CreateOrderNotFound{
				Code:    http.StatusNotFound,
				Message: err.Error(),
			}, nil

		case errors.Is(err, errs.ErrOutOfStock):
			return &orderv1.CreateOrderConflict{
				Code:    http.StatusConflict,
				Message: err.Error(),
			}, nil

		default:
			return &orderv1.CreateOrderInternalServerError{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
			}, nil
		}
	}

	// 8. Вернуть order_uuid и total_price
	return converter.OrderCreateRsToDTO(order), nil
}

package v1

import (
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

// OrderHandler реализует интерфейс orderv1.Handler, сгенерированный ogen.
type orderHandler struct {
	orderv1.UnimplementedHandler

	orderService OrderService
}

// NewOrderHandler создаёт новый обработчик заказов.
func NewOrderHandler(orderService OrderService) *orderHandler {
	return &orderHandler{orderService: orderService}
}

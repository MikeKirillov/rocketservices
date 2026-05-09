package converter

import (
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func OrderToModel(order record.Order, items []record.OrderItem) model.Order {
	result := model.Order{
		OrderUUID: order.UUID,
		Status:    model.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt,
	}

	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}

	if order.PaymentMethod != nil {
		method := model.PaymentMethod(*order.PaymentMethod)
		result.PaymentMethod = &method
	}

	modelItems := make([]model.OrderItem, 0, len(items))
	var totalPrice int64
	for _, i := range items {
		p := model.OrderItem{
			Uuid:  i.PartUUID,
			Type:  model.PartType(i.PartType),
			Name:  &i.PartType,
			Price: i.Price,
		}

		modelItems = append(modelItems, p)
		totalPrice += i.Price
	}
	result.Items = modelItems
	result.TotalPrice = totalPrice

	return result
}

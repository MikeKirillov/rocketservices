package converter

import (
	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func OrderToRecord(order model.Order) record.Order {
	result := record.Order{
		UUID:      order.OrderUUID,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt,
		Version:   order.Version,
	}

	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}
	if order.PaymentMethod != nil {
		result.PaymentMethod = (*string)(order.PaymentMethod)
	}
	if order.UpdatedAt != nil {
		result.UpdatedAt = order.UpdatedAt
	}

	return result
}

func OrderItemToRecord(partId, orderId uuid.UUID, item model.OrderItem) record.OrderItem {
	return record.OrderItem{
		UUID:      partId,
		OrderUUID: orderId,
		PartUUID:  item.Uuid,
		PartType:  string(item.Type),
		Price:     item.Price,
	}
}

func OrderToModel(order record.Order) model.Order {
	result := model.Order{
		OrderUUID: order.UUID,
		Status:    model.OrderStatus(order.Status),
		CreatedAt: order.CreatedAt,
		Version:   order.Version,
	}

	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}

	if order.PaymentMethod != nil {
		method := model.PaymentMethod(*order.PaymentMethod)
		result.PaymentMethod = &method
	}

	if order.UpdatedAt != nil {
		result.UpdatedAt = order.UpdatedAt
	}

	modelItems := make([]model.OrderItem, 0, len(order.Items))
	var totalPrice int64
	for _, i := range order.Items {
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

package converter

import (
	"time"

	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func OrderToModel(order record.Order) model.Order {
	result := model.Order{
		OrderUUID:  order.OrderUUID,
		HullUUID:   order.HullUUID,
		EngineUUID: order.EngineUUID,
		TotalPrice: order.TotalPrice,
		Status:     model.OrderStatus(order.Status),
		CreatedAt:  order.CreatedAt,
	}
	if order.ShieldUUID != nil {
		result.ShieldUUID = order.ShieldUUID
	}
	if order.WeaponUUID != nil {
		result.WeaponUUID = order.WeaponUUID
	}
	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}
	if order.PaymentMethod != nil {
		payment := model.PaymentMethod(*order.PaymentMethod)
		result.PaymentMethod = &payment
	}
	return result
}

func OrderToRepoCreate(
	order model.CreateOrderRq,
	totalPrice int64,
	orderUUID string,
	status model.OrderStatus,
) record.Order {
	result := record.Order{
		OrderUUID:  orderUUID,
		HullUUID:   order.HullUUID,
		EngineUUID: order.EngineUUID,
		TotalPrice: totalPrice,
		Status:     string(status),
		CreatedAt:  time.Now(),
	}
	if order.ShieldUUID != nil {
		result.ShieldUUID = order.ShieldUUID
	}
	if order.WeaponUUID != nil {
		result.WeaponUUID = order.WeaponUUID
	}
	return result
}

func OrderToRepoUpdate(
	order model.Order,
	status model.OrderStatus,
	transactionUUID *string,
	paymentMethod *model.PaymentMethod,
) record.Order {
	result := record.Order{
		OrderUUID:  order.OrderUUID,
		HullUUID:   order.HullUUID,
		EngineUUID: order.EngineUUID,
		TotalPrice: order.TotalPrice,
		Status:     string(status),
		CreatedAt:  order.CreatedAt,
	}
	if order.ShieldUUID != nil {
		result.ShieldUUID = order.ShieldUUID
	}
	if order.WeaponUUID != nil {
		result.WeaponUUID = order.WeaponUUID
	}
	if transactionUUID != nil {
		result.TransactionUUID = transactionUUID
	}
	if paymentMethod != nil {
		payment := string(*paymentMethod)
		result.PaymentMethod = &payment
	}

	return result
}

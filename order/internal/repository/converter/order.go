package converter

import (
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	"github.com/MikeKirillov/rocketservices/order/internal/repository/record"
)

func OrderToModel(order record.Order) model.Order {
	result := model.Order{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
		Status:     model.OrderStatus(order.Status),
		CreatedAt:  order.CreatedAt,
	}
	parts := []model.Part{
		{
			Uuid: order.HullUUID,
			Type: model.PartTypeHull,
		},
		{
			Uuid: order.EngineUUID,
			Type: model.PartTypeEngine,
		},
	}
	if order.ShieldUUID != nil {
		parts = append(parts, model.Part{
			Uuid: *order.ShieldUUID,
			Type: model.PartTypeShield,
		})
	}
	if order.WeaponUUID != nil {
		parts = append(parts, model.Part{
			Uuid: *order.WeaponUUID,
			Type: model.PartTypeWeapon,
		})
	}
	result.Parts = parts

	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}
	if order.PaymentMethod != nil {
		payment := model.PaymentMethod(*order.PaymentMethod)
		result.PaymentMethod = &payment
	}
	return result
}

func OrderToRepo(order model.Order) record.Order {
	result := record.Order{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
		Status:     string(order.Status),
		CreatedAt:  order.CreatedAt,
	}

	for _, p := range order.Parts {
		switch p.Type {
		case model.PartTypeHull:
			result.HullUUID = p.Uuid
		case model.PartTypeEngine:
			result.EngineUUID = p.Uuid
		case model.PartTypeShield:
			result.ShieldUUID = &p.Uuid
		case model.PartTypeWeapon:
			result.WeaponUUID = &p.Uuid
		}
	}

	if order.TransactionUUID != nil {
		result.TransactionUUID = order.TransactionUUID
	}
	if order.PaymentMethod != nil {
		payment := string(*order.PaymentMethod)
		result.PaymentMethod = &payment
	}
	return result
}

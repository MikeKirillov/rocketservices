package converter

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/api/input"
	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func OrderToDTO(order model.Order) (*orderv1.OrderDto, error) {
	result := &orderv1.OrderDto{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
		Status:     orderv1.OrderStatus(order.Status),
		CreatedAt:  order.CreatedAt,
		UserUUID:   order.UserUUID,
	}
	for _, p := range order.Items {
		switch p.Type {
		case model.PartTypeHull:
			result.SetHullUUID(p.Uuid)
		case model.PartTypeEngine:
			result.SetEngineUUID(p.Uuid)
		case model.PartTypeShield:
			result.SetShieldUUID(orderv1.NewOptNilUUID(p.Uuid))
		case model.PartTypeWeapon:
			result.SetWeaponUUID(orderv1.NewOptNilUUID(p.Uuid))
		}
	}

	if order.TransactionUUID != nil {
		result.SetTransactionUUID(orderv1.NewOptNilUUID(*order.TransactionUUID))
	}

	if order.PaymentMethod != nil {
		result.SetPaymentMethod(orderv1.NewOptNilPaymentMethod(orderv1.PaymentMethod(*order.PaymentMethod)))
	}

	return result, nil
}

func OrderCreateRqToModel(orderRq *orderv1.CreateOrderRequest) (input.CreateOrderRq, error) {
	// Валидация: hull_uuid и engine_uuid обязательны
	if err := checkRequiredField(orderRq.HullUUID, "hull_uuid"); err != nil {
		return input.CreateOrderRq{}, err
	}
	if err := checkRequiredField(orderRq.EngineUUID, "engine_uuid"); err != nil {
		return input.CreateOrderRq{}, err
	}

	items := []model.OrderItem{
		{
			Uuid: orderRq.HullUUID,
			Type: model.PartTypeHull,
		},
		{
			Uuid: orderRq.EngineUUID,
			Type: model.PartTypeEngine,
		},
	}

	if v, ok := orderRq.GetShieldUUID().Get(); ok {
		items = append(items, model.OrderItem{Uuid: v, Type: model.PartTypeShield})
	}
	if v, ok := orderRq.GetWeaponUUID().Get(); ok {
		items = append(items, model.OrderItem{Uuid: v, Type: model.PartTypeWeapon})
	}
	return input.CreateOrderRq{Items: items}, nil
}

func checkRequiredField(field uuid.UUID, fieldName string) error {
	if field == uuid.Nil {
		slog.Error("некорректный запрос",
			slog.String("не передано обязательное поле", fieldName),
		)
		return errs.ErrOmittedRequiredField
	}
	return nil
}

func OrderCreateRsToDTO(order input.CreateOrderRs) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
	}
}

func PayOrderRqToModel(id uuid.UUID, paymentMethod orderv1.PaymentMethod) (input.PayOrderRq, error) {
	payment, err := model.NewPaymentMethod(string(paymentMethod))
	if err != nil {
		return input.PayOrderRq{}, err
	}
	return input.PayOrderRq{
		OrderUUID:     id,
		PaymentMethod: payment,
	}, nil
}

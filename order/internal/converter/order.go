package converter

import (
	"fmt"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

func OrderToDTO(order model.Order) (*orderv1.OrderDto, error) {
	orderUUID, err := uuid.Parse(order.OrderUUID)
	if err != nil {
		return &orderv1.OrderDto{}, fmt.Errorf("поле OrderUUID: %w", errs.ErrInvalidUUID)
	}

	hullUUID, err := uuid.Parse(order.HullUUID)
	if err != nil {
		return &orderv1.OrderDto{}, fmt.Errorf("поле HullUUID: %w", errs.ErrInvalidUUID)
	}

	engineUUID, err := uuid.Parse(order.EngineUUID)
	if err != nil {
		return &orderv1.OrderDto{}, fmt.Errorf("поле EngineUUID: %w", errs.ErrInvalidUUID)
	}

	var shieldUUID orderv1.OptNilUUID
	if order.ShieldUUID != nil {
		id, err := uuid.Parse(*order.ShieldUUID)
		if err != nil {
			return &orderv1.OrderDto{}, fmt.Errorf("поле ShieldUUID: %w", errs.ErrInvalidUUID)
		}
		shieldUUID = orderv1.NewOptNilUUID(id)
	}

	var weaponUUID orderv1.OptNilUUID
	if order.WeaponUUID != nil {
		id, err := uuid.Parse(*order.WeaponUUID)
		if err != nil {
			return &orderv1.OrderDto{}, fmt.Errorf("поле WeaponUUID: %w", errs.ErrInvalidUUID)
		}
		weaponUUID = orderv1.NewOptNilUUID(id)
	}

	var transactionUUID orderv1.OptNilUUID
	if order.TransactionUUID != nil {
		id, err := uuid.Parse(*order.TransactionUUID)
		if err != nil {
			return &orderv1.OrderDto{}, fmt.Errorf("поле TransactionUUID: %w", errs.ErrInvalidUUID)
		}
		transactionUUID = orderv1.NewOptNilUUID(id)
	}

	var paymentMethod orderv1.OptNilPaymentMethod
	if order.PaymentMethod != nil {
		paymentMethod = orderv1.NewOptNilPaymentMethod(orderv1.PaymentMethod(*order.PaymentMethod))
	}

	return &orderv1.OrderDto{
		OrderUUID:       orderUUID,
		HullUUID:        hullUUID,
		EngineUUID:      engineUUID,
		ShieldUUID:      shieldUUID,
		WeaponUUID:      weaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: transactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          orderv1.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}, nil
}

func OrderCreateRqToModel(orderRq *orderv1.CreateOrderRequest) model.CreateOrderRq {
	result := model.CreateOrderRq{
		HullUUID:   orderRq.HullUUID.String(),
		EngineUUID: orderRq.EngineUUID.String(),
	}
	if orderRq.GetShieldUUID().IsSet() && !orderRq.GetShieldUUID().IsNull() {
		val := orderRq.GetShieldUUID().Value.String()
		result.ShieldUUID = &val
	}
	if orderRq.GetWeaponUUID().IsSet() && !orderRq.GetWeaponUUID().IsNull() {
		val := orderRq.GetWeaponUUID().Value.String()
		result.WeaponUUID = &val
	}
	return result
}

func OrderCreateRsToDTO(order model.CreateOrderRs) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{
		OrderUUID:  order.OrderUUID,
		TotalPrice: order.TotalPrice,
	}
}

func PayOrderRqToModel(id uuid.UUID, paymentMethod orderv1.PaymentMethod) (model.PayOrderRq, error) {
	payment, err := paymentMethodToModel(paymentMethod)
	if err != nil {
		return model.PayOrderRq{}, err
	}
	return model.PayOrderRq{
		OrderUUID:     id.String(),
		PaymentMethod: payment,
	}, nil
}

func paymentMethodToModel(payment orderv1.PaymentMethod) (model.PaymentMethod, error) {
	switch payment {
	case orderv1.PaymentMethodCARD:
		return model.PaymentMethodCard, nil
	case orderv1.PaymentMethodSBP:
		return model.PaymentMethodSbp, nil
	case orderv1.PaymentMethodCREDITCARD:
		return model.PaymentMethodCreditCard, nil
	case orderv1.PaymentMethodINVESTORMONEY:
		return model.PaymentMethodInvestorMoney, nil
	default:
		return model.PaymentMethodUnspecified, errs.ErrInvalidPaymentMethod
	}
}

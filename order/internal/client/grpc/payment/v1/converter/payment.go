package converter

import (
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func PaymentToProto(orderRq model.PayOrderRq) *paymentv1.PayOrderRequest {
	return &paymentv1.PayOrderRequest{
		OrderUuid:     orderRq.OrderUUID,
		PaymentMethod: paymentMethodToProto(orderRq.PaymentMethod),
	}
}

func paymentMethodToProto(payment model.PaymentMethod) paymentv1.PaymentMethod {
	switch payment {
	case model.PaymentMethodCard:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CARD

	case model.PaymentMethodSbp:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_SBP

	case model.PaymentMethodCreditCard:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD

	default:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	}
}

package converter

import (
	"github.com/MikeKirillov/rocketservices/payment/internal/model"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func OrderToModel(id string, paymentMethod paymentv1.PaymentMethod) model.Order {
	return model.Order{
		Uuid:          id,
		PaymentMethod: PayMethodToModel(paymentMethod),
	}
}

func PayMethodToModel(paymentMethod paymentv1.PaymentMethod) model.PaymentMethod {
	switch paymentMethod {
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CARD:
		return model.PaymentMethodCard

	case paymentv1.PaymentMethod_PAYMENT_METHOD_SBP:
		return model.PaymentMethodSbp

	case paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return model.PaymentMethodCreditCard

	case paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY:
		return model.PaymentMethodInvestorMoney

	default:
		return model.PaymentMethodUnspecified
	}
}

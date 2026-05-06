package converter

import (
	"fmt"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/payment/internal/errors"
	"github.com/MikeKirillov/rocketservices/payment/internal/model"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func PaymentToModel(order *paymentv1.PayOrderRequest) (model.Payment, error) {
	id, err := uuid.Parse(order.GetOrderUuid())
	if err != nil {
		return model.Payment{}, errs.ErrPaymentInvalidUUID
	}
	pm, err := PayMethodToModel(order.GetPaymentMethod())
	if err != nil {
		return model.Payment{}, err
	}
	return model.Payment{
		Uuid:          id,
		PaymentMethod: pm,
	}, nil
}

func PayMethodToModel(paymentMethod paymentv1.PaymentMethod) (model.PaymentMethod, error) {
	switch paymentMethod {
	case paymentv1.PaymentMethod_PAYMENT_METHOD_CARD:
		return model.PaymentMethodCard, nil

	case paymentv1.PaymentMethod_PAYMENT_METHOD_SBP:
		return model.PaymentMethodSbp, nil

	case paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD:
		return model.PaymentMethodCreditCard, nil

	case paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY:
		return model.PaymentMethodInvestorMoney, nil

	default:
		return "", fmt.Errorf("способ оплаты %s: %w", paymentMethod, errs.ErrPaymentInvalidPayMethod)
	}
}

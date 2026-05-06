package model

import (
	"github.com/google/uuid"
)

type Payment struct {
	Uuid          uuid.UUID
	PaymentMethod PaymentMethod
}

type PaymentMethod string

const (
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSbp           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

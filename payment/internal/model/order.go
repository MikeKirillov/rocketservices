package model

type Order struct {
	Uuid          string
	PaymentMethod PaymentMethod
}

type PaymentMethod string

const (
	PaymentMethodUnspecified   PaymentMethod = "UNSPECIFIED"
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSbp           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCard, PaymentMethodSbp, PaymentMethodCreditCard, PaymentMethodInvestorMoney:
		return true
	default:
		return false
	}
}

package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
)

type PaymentMethod string

const (
	PaymentMethodUnspecified   PaymentMethod = "UNSPECIFIED"
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSbp           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

// Order представляет заказ на постройку космического корабля.
type Order struct {
	OrderUUID       string
	HullUUID        string
	EngineUUID      string
	ShieldUUID      *string // опциональный
	WeaponUUID      *string // опциональный
	TotalPrice      int64   // в копейках
	TransactionUUID *string
	PaymentMethod   *PaymentMethod
	Status          OrderStatus
	CreatedAt       time.Time
}

type CreateOrderRq struct {
	HullUUID   string
	EngineUUID string
	ShieldUUID *string
	WeaponUUID *string
}

func (r *CreateOrderRq) PartUUIDsAsStringSlice() []string {
	uuids := []string{r.HullUUID, r.EngineUUID}
	if r.ShieldUUID != nil {
		uuids = append(uuids, *r.ShieldUUID)
	}
	if r.WeaponUUID != nil {
		uuids = append(uuids, *r.WeaponUUID)
	}
	return uuids
}

type CreateOrderRs struct {
	OrderUUID  uuid.UUID
	TotalPrice int64 // в копейках
}

type PayOrderRq struct {
	OrderUUID     string
	PaymentMethod PaymentMethod
}

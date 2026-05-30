package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/order/internal/errors"
)

type OrderStatus string

const (
	OrderStatusPendingPayment OrderStatus = "PENDING_PAYMENT"
	OrderStatusPaid           OrderStatus = "PAID"
	OrderStatusCancelled      OrderStatus = "CANCELLED"
	OrderStatusAssembled      OrderStatus = "ASSEMBLED"
)

type PaymentMethod string

const (
	PaymentMethodCard          PaymentMethod = "CARD"
	PaymentMethodSbp           PaymentMethod = "SBP"
	PaymentMethodCreditCard    PaymentMethod = "CREDIT_CARD"
	PaymentMethodInvestorMoney PaymentMethod = "INVESTOR_MONEY"
)

func NewPaymentMethod(s string) (PaymentMethod, error) {
	pm := PaymentMethod(s)
	switch pm {
	case PaymentMethodCard, PaymentMethodSbp, PaymentMethodCreditCard, PaymentMethodInvestorMoney:
		return pm, nil
	default:
		return "", fmt.Errorf("способ оплаты %s: %w", s, errs.ErrInvalidPaymentMethod)
	}
}

// Order представляет заказ на постройку космического корабля.
type Order struct {
	OrderUUID       uuid.UUID
	Items           []OrderItem
	TotalPrice      int64 // в копейках
	TransactionUUID *uuid.UUID
	PaymentMethod   *PaymentMethod
	Status          OrderStatus
	CreatedAt       time.Time
	UpdatedAt       *time.Time
	Version         int64
	UserUUID        uuid.UUID
}

func NewOrderToCreate(
	parts []OrderItem,
	totalPrice int64,
	userUUID uuid.UUID,
) Order {
	return Order{
		OrderUUID:  uuid.New(),
		Items:      parts,
		TotalPrice: totalPrice,
		Status:     OrderStatusPendingPayment,
		CreatedAt:  time.Now(),
		UserUUID:   userUUID,
	}
}

// CanModify() позволяет менять заказы только в статусе PENDING_PAYMENT.
func (o *Order) CanModify() error {
	if o.Status == OrderStatusCancelled {
		return errs.ErrOrderAlreadyCancelled
	}
	if o.Status == OrderStatusPaid {
		return errs.ErrOrderAlreadyPaid
	}
	if o.Status == OrderStatusAssembled {
		return errs.ErrOrderAlreadyAssembled
	}
	if o.Status != OrderStatusPendingPayment {
		return errs.ErrOrderUnknownStatus
	}
	return nil
}

// CanCommit() позволяет производить только заказы в статусе PAID.
func (o *Order) CanCommit() error {
	if o.Status == OrderStatusCancelled {
		return errs.ErrOrderAlreadyCancelled
	}
	if o.Status == OrderStatusAssembled {
		return errs.ErrOrderAlreadyAssembled
	}
	if o.Status == OrderStatusPendingPayment {
		return errs.ErrOrderUnpaid
	}
	return nil
}

// Pay() меняет статус заказа на PAID, устанавливает способ оплаты и идентификатор транзакции.
func (o *Order) Pay(transaction uuid.UUID, payment *PaymentMethod) {
	if payment != nil {
		o.PaymentMethod = payment
	}
	o.TransactionUUID = &transaction
	o.Status = OrderStatusPaid
	o.UpdatedAt = new(time.Now())
}

// Cancel() меняет статус заказа на CANCELLED.
func (o *Order) Cancel() error {
	err := o.CanModify()
	if err != nil {
		return err
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = new(time.Now())
	return nil
}

// Commit() меняет статус заказа на ASSEMBLED.
func (o *Order) Commit() {
	o.Status = OrderStatusAssembled
	o.UpdatedAt = new(time.Now())
}

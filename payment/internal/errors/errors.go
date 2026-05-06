package errs

import "errors"

var (
	ErrPaymentOmittedUUID      = errors.New("order_uuid обязателен")
	ErrPaymentInvalidUUID      = errors.New("неверный формат uuid")
	ErrPaymentInvalidPayMethod = errors.New("некорректный payment_method")
)

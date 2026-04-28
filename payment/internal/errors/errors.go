package errs

import "errors"

var (
	ErrOrderOmittedUUID      = errors.New("order_uuid обязателен")
	ErrOrderInvalidUUID      = errors.New("неверный формат uuid")
	ErrOrderInvalidPayMethod = errors.New("некорректный payment_method")
)

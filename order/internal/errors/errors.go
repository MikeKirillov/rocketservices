package errs

import "errors"

var (
	// Ошибки заказов.
	ErrOrderNotFound               = errors.New("заказ не найден")
	ErrOrderAlreadyPaidOrCancalled = errors.New("заказ уже оплачен или был отменен")
	ErrOrderCancelled              = errors.New("заказ отменён")
	ErrOrderCancelling             = errors.New("невозможно отменить заказ")

	// Ошибки деталей.
	ErrPartNotFound = errors.New("деталь не найдена")
	ErrOutOfStock   = errors.New("деталь отсутствует на складе")

	// Ошибки валидации.
	ErrOmittedRequiredField = errors.New("в запросе не передано обязательное поле")
	ErrInvalidUUID          = errors.New("неверный формат UUID")
	ErrInvalidPaymentMethod = errors.New("неизвестный способ оплаты")

	// Ошибки интеграционных сервисов.
	ErrGrpcInternal        = errors.New("внутренняя ошибка сервера")
	ErrGrpcTimeout         = errors.New("таймаут grpc сервера")
	ErrGrpcInvalidArgument = errors.New("некорректный запрос на grpc сервер")
	ErrGrpcUnavailable     = errors.New("сервис недоступен")
	ErrInventoryNotFound   = errors.New("детали не найдены")
)

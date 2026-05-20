package errs

import "errors"

var (
	// Ошибки заказов.
	ErrOrderNotFound         = errors.New("заказ не найден")
	ErrOrderConflict         = errors.New("заказ не удалось обновить")
	ErrOrderAlreadyPaid      = errors.New("заказ уже оплачен")
	ErrOrderAlreadyCancelled = errors.New("заказ уже отменён")
	ErrOrderUnknownStatus    = errors.New("неизвестный статус заказа")
	ErrOrderCannotCancel     = errors.New("невозможно отменить заказ")

	// Ошибки деталей.
	ErrPartNotFound  = errors.New("деталь не найдена")
	ErrOutOfStock    = errors.New("деталь отсутствует на складе")
	ErrWrongPartType = errors.New("неизвестный тип детали")

	// Ошибки валидации.
	ErrOmittedRequiredField = errors.New("в запросе не передано обязательное поле")
	ErrInvalidUUID          = errors.New("неверный формат UUID")
	ErrInvalidPaymentMethod = errors.New("неизвестный способ оплаты")
	ErrUnknown              = errors.New("неизвестная ошибка")

	// Ошибки интеграционных сервисов.
	ErrGrpcInternal        = errors.New("внутренняя ошибка сервера")
	ErrGrpcTimeout         = errors.New("таймаут grpc сервера")
	ErrGrpcInvalidArgument = errors.New("некорректный запрос на grpc сервер")
	ErrGrpcUnavailable     = errors.New("сервис недоступен")
	ErrGrpcNotFound        = errors.New("детали не найдены")
)

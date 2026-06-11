package errs

import "errors"

var (
	ErrPartOmittedUUID   = errors.New("uuid обязателен")
	ErrPartInvalidUUID   = errors.New("неверный формат uuid")
	ErrPartNotFound      = errors.New("деталь не найдена")
	ErrWrongPartType     = errors.New("неизвестный тип детали")
	ErrInvalidProperties = errors.New("некорректные свойства компонента")
	ErrOutOfStock        = errors.New("деталь отсутствует на складе")
	ErrNothingToRelease  = errors.New("резерв равен нулю")
	ErrNothingToCommit   = errors.New("резерв равен нулю или деталь отсутствует на складе")
	ErrInvalidUUID       = errors.New("неверный формат UUID")
	ErrIncompatibleParts = errors.New("детали несовместимы")
	ErrPartTypeMismatch  = errors.New("тип детали не соответствует слоту корабля")

	// Ошибки интеграционных сервисов.
	ErrGrpcInternal           = errors.New("внутренняя ошибка сервера")
	ErrGrpcTimeout            = errors.New("таймаут grpc сервера")
	ErrGrpcInvalidArgument    = errors.New("некорректный запрос на grpc сервер")
	ErrGrpcUnavailable        = errors.New("сервис недоступен")
	ErrGrpcNotFound           = errors.New("детали не найдены")
	ErrGrpcCancelledOperation = errors.New("операция отклонена")
	ErrGrpcInvalidSession     = errors.New("недействительная сессия")
)

package errs

import "errors"

var (
	ErrPartOmittedUUID   = errors.New("uuid обязателен")
	ErrPartInvalidUUID   = errors.New("неверный формат uuid")
	ErrPartNotFound      = errors.New("деталь не найдена")
	ErrWrongPartType     = errors.New("неизвестный тип детали")
	ErrInvalidProperties = errors.New("некорректные свойства компонента")
	ErrOutOfStock        = errors.New("нет свободных единиц на складе")
	ErrNothingToRelease  = errors.New("резерв равен нулю")
)

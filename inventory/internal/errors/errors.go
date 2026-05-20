package errs

import "errors"

var (
	ErrPartOmittedUUID = errors.New("uuid обязателен")
	ErrPartInvalidUUID = errors.New("неверный формат uuid")
	ErrPartNotFound    = errors.New("деталь не найдена")
	ErrWrongPartType   = errors.New("неизвестный тип детали")
)

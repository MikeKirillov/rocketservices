package part

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Reserve(ctx context.Context, ids []uuid.UUID) error {
	return s.txManager.Do(ctx, func(ctx context.Context) error {
		// Загрузите детали по uuids через репозиторий.
		parts, err := s.partRepository.ListByUUIDsForUpdate(ctx, ids)
		if err != nil {
			switch {
			case errors.Is(err, errs.ErrPartNotFound):
				return sharedErrs.NewNotFound(err)
			case errors.Is(err, errs.ErrWrongPartType):
				return sharedErrs.NewConflict(err)
			default:
				return sharedErrs.NewInternal(err)
			}
		}
		// Вызовите соответствующий метод entity (Reserve() / Release()) для каждой детали —
		// entity сама вернёт ошибку при нарушении инварианта.
		for i := range parts {
			if err = parts[i].Reserve(); err != nil {
				return sharedErrs.NewResourceExhausted(err)
			}
		}
		// Сохраните обновлённое состояние через UpdateReservedBatch — один SQL-запрос для всех деталей.
		if err = s.partRepository.UpdateReservedBatch(ctx, parts); err != nil {
			return sharedErrs.NewInternal(err)
		}
		return nil
	})
}

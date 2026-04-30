package part

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id string) (model.Part, error) {
	// Проверить, что uuid не пустой → INVALID_ARGUMENT
	if id == "" {
		return model.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartOmittedUUID)
	}

	// Валидировать формат UUID → INVALID_ARGUMENT. Валидация формата UUID v4
	// Можно использовать github.com/google/uuid:
	if _, err := uuid.Parse(id); err != nil {
		return model.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartInvalidUUID)
	}

	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		return model.Part{}, sharedErrs.NewNotFound(err)
	}
	return part, nil
}

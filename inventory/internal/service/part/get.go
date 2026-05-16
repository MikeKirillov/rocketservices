package part

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model/entity"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (entity.Part, error) {
	// Проверить, что uuid не пустой → INVALID_ARGUMENT
	if id == uuid.Nil {
		return entity.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartOmittedUUID)
	}
	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrPartNotFound) {
			return entity.Part{}, sharedErrs.NewNotFound(err)
		}
		return entity.Part{}, sharedErrs.NewInternal(err)
	}
	return part, nil
}

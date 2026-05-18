package part

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	"github.com/MikeKirillov/rocketservices/inventory/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Get(ctx context.Context, id uuid.UUID) (model.Part, error) {
	// Проверить, что uuid не пустой → INVALID_ARGUMENT
	if id == uuid.Nil {
		return model.Part{}, sharedErrs.NewInvalidArgument(errs.ErrPartOmittedUUID)
	}
	part, err := s.partRepository.Get(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrPartNotFound) {
			return model.Part{}, sharedErrs.NewNotFound(err)
		}
		return model.Part{}, sharedErrs.NewInternal(err)
	}
	return part, nil
}

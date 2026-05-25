package part

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/inventory/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Commit(ctx context.Context, ids []uuid.UUID) error {
	return s.txManager.Do(ctx, func(ctx context.Context) error {
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

		for i := range parts {
			if err = parts[i].Commit(); err != nil {
				return sharedErrs.NewFailedPrecondition(err)
			}
		}

		if err = s.partRepository.UpdateCommitBatch(ctx, parts); err != nil {
			return sharedErrs.NewInternal(err)
		}
		return nil
	})
}

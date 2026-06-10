package iam

import (
	"context"

	"github.com/google/uuid"

	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Logout(ctx context.Context, session uuid.UUID) error {
	if err := s.sessionRepository.Delete(ctx, session); err != nil {
		return sharedErrs.NewInternal(err)
	}
	return nil
}

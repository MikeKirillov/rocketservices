package iam

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	user, err := s.userRepository.GetUserByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return model.User{}, sharedErrs.NewNotFound(err)
		}
		return model.User{}, sharedErrs.NewInternal(err)
	}

	return user, nil
}

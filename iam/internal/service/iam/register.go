package iam

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Register(ctx context.Context, input input.RegisterInput) (uuid.UUID, error) {
	user, err := model.NewUser(input.Login, input.Password)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidLogin) || errors.Is(err, errs.ErrWeakPassword) {
			return uuid.Nil, sharedErrs.NewInvalidArgument(err)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	err = s.userRepository.InsertUser(ctx, user)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			return uuid.Nil, sharedErrs.NewConflict(err)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	return user.UUID, nil
}

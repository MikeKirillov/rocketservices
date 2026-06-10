package iam

import (
	"context"
	"errors"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Whoami(ctx context.Context, sessionUUID uuid.UUID) (model.User, model.Session, error) {
	session, err := s.sessionRepository.Get(ctx, sessionUUID)
	if err != nil {
		return model.User{}, model.Session{}, err
	}

	user, err := s.userRepository.GetUserByUUID(ctx, session.UserUUID)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return model.User{}, model.Session{}, sharedErrs.NewNotFound(err)
		}
		return model.User{}, model.Session{}, sharedErrs.NewInternal(err)
	}

	return user, session, nil
}

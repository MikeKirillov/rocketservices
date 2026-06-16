package iam

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
)

func (s *service) Whoami(ctx context.Context, sessionUUID uuid.UUID) (model.User, model.Session, error) {
	session, user, err := s.sessionRepository.Get(ctx, sessionUUID)
	if err != nil {
		return model.User{}, model.Session{}, err
	}

	return user, session, nil
}

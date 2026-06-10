package iam

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *service) Login(ctx context.Context, input input.LoginPassInput) (uuid.UUID, error) {
	// Проверка заполненности логина и пароля
	if input.Login == "" || input.Password == "" {
		return uuid.Nil, sharedErrs.NewInvalidArgument(errs.ErrEmptyCredential)
	}

	user, err := s.userRepository.GetUserByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return uuid.Nil, sharedErrs.NewNotFound(errs.ErrInvalidCredentials)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	// Проверка корректности пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return uuid.Nil, sharedErrs.NewUnauthorized(errs.ErrInvalidCredentials)
	}

	session := model.NewSession(user)
	if err := s.sessionRepository.Set(ctx, session); err != nil {
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	return session.UUID, nil
}

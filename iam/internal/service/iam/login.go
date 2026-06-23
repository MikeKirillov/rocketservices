package iam

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *Service) Login(ctx context.Context, input input.LoginPassInput) (uuid.UUID, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.Login", trace.WithAttributes(
		attribute.String("user.login", input.Login)),
	)
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "iam.Login.Ok")
		}
		span.End()
	}()

	// Проверка заполненности логина и пароля
	if input.Login == "" || input.Password == "" {
		resultErr = sharedErrs.NewInvalidArgument(errs.ErrEmptyCredential)
		return uuid.Nil, resultErr
	}

	user, err := s.userRepository.GetUserByLogin(ctx, input.Login)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			resultErr = sharedErrs.NewUnauthorized(errs.ErrInvalidCredentials)
			return uuid.Nil, resultErr
		}
		resultErr = sharedErrs.NewInternal(err)
		return uuid.Nil, resultErr
	}

	// Проверка корректности пароля
	if err := user.Password.Verify(input.Password); err != nil {
		resultErr = sharedErrs.NewUnauthorized(errs.ErrInvalidCredentials)
		return uuid.Nil, resultErr
	}

	session := model.NewSession(s.ttl)
	if err := s.sessionRepository.Set(ctx, session, user, s.ttl); err != nil {
		resultErr = sharedErrs.NewInternal(err)
		return uuid.Nil, resultErr
	}

	span.SetAttributes(
		attribute.String("session.uuid", session.UUID.String()),
	)

	return session.UUID, nil
}

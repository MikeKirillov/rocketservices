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
	defer span.End()

	// Проверка заполненности логина и пароля
	if input.Login == "" || input.Password == "" {
		span.RecordError(errs.ErrEmptyCredential)
		span.SetStatus(codes.Error, errs.ErrEmptyCredential.Error())

		return uuid.Nil, sharedErrs.NewInvalidArgument(errs.ErrEmptyCredential)
	}

	user, err := s.userRepository.GetUserByLogin(ctx, input.Login)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		if errors.Is(err, errs.ErrUserNotFound) {
			return uuid.Nil, sharedErrs.NewUnauthorized(errs.ErrInvalidCredentials)
		}
		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	// Проверка корректности пароля
	if err := user.Password.Verify(input.Password); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return uuid.Nil, sharedErrs.NewUnauthorized(errs.ErrInvalidCredentials)
	}

	session := model.NewSession(s.ttl)
	if err := s.sessionRepository.Set(ctx, session, user, s.ttl); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return uuid.Nil, sharedErrs.NewInternal(err)
	}

	span.SetAttributes(
		attribute.String("session.uuid", session.UUID.String()),
	)

	return session.UUID, nil
}

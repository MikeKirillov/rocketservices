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

func (s *Service) Register(ctx context.Context, input input.LoginPassInput) (uuid.UUID, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.Register", trace.WithAttributes(
		attribute.String("user.login", input.Login),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "iam.Register.Ok")
		}
		span.End()
	}()

	user, err := model.NewUser(input.Login, input.Password)
	if err != nil {
		if errors.Is(err, errs.ErrInvalidLogin) || errors.Is(err, errs.ErrWeakPassword) {
			resultErr = sharedErrs.NewInvalidArgument(err)
			return uuid.Nil, resultErr
		}
		resultErr = sharedErrs.NewInternal(err)
		return uuid.Nil, resultErr
	}

	err = s.userRepository.InsertUser(ctx, user)
	if err != nil {
		if errors.Is(err, errs.ErrUserAlreadyExists) {
			resultErr = sharedErrs.NewConflict(err)
			return uuid.Nil, resultErr
		}
		resultErr = sharedErrs.NewInternal(err)
		return uuid.Nil, resultErr
	}

	span.SetAttributes(attribute.String("user.uuid", user.UUID.String()))

	return user.UUID, nil
}

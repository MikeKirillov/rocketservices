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
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.GetUser", trace.WithAttributes(
		attribute.String("user.uuid", id.String()),
	))
	defer func() {
		span.SetStatus(codes.Ok, "iam.GetUser.Ok")
		span.End()
	}()

	user, err := s.userRepository.GetUserByUUID(ctx, id)
	if err != nil {
		if errors.Is(err, errs.ErrUserNotFound) {
			return model.User{}, sharedErrs.NewNotFound(err)
		}
		return model.User{}, sharedErrs.NewInternal(err)
	}

	return user, nil
}

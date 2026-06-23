package iam

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
)

func (s *Service) Whoami(ctx context.Context, sessionUUID uuid.UUID) (model.User, model.Session, error) {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.Whoami", trace.WithAttributes(
		attribute.String("session.uuid", sessionUUID.String()),
	))
	var resultErr error
	defer func() {
		if resultErr != nil {
			span.RecordError(resultErr)
			span.SetStatus(codes.Error, resultErr.Error())
		} else {
			span.SetStatus(codes.Ok, "iam.Whoami.Ok")
		}
		span.End()
	}()

	session, user, err := s.sessionRepository.Get(ctx, sessionUUID)
	if err != nil {
		resultErr = err
		return model.User{}, model.Session{}, resultErr
	}

	span.SetAttributes(
		attribute.String("user.uuid", user.UUID.String()),
		attribute.String("user.login", user.Login),
		attribute.String("session.uuid", sessionUUID.String()),
	)

	return user, session, nil
}

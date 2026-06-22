package iam

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

func (s *Service) Logout(ctx context.Context, session uuid.UUID) error {
	// Подключение трейсинга бизнес-логики
	ctx, span := otel.Tracer("iam-service").Start(ctx, "iam.Logout", trace.WithAttributes(
		attribute.String("session.uuid", session.String()),
	))
	defer span.End()

	if err := s.sessionRepository.Delete(ctx, session); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return sharedErrs.NewInternal(err)
	}
	return nil
}

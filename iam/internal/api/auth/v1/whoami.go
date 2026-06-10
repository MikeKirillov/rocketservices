package v1

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/api/converter"
	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Whoami(ctx context.Context, req *authv1.WhoamiRequest) (*authv1.WhoamiResponse, error) {
	if req.GetSessionUuid() == "" {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrEmptySessionID)
	}
	sessionUUID, err := uuid.Parse(req.GetSessionUuid())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrInvalidUUID)
	}

	user, session, err := a.authService.Whoami(ctx, sessionUUID)
	if err != nil {
		if errors.Is(err, errs.ErrSessionNotFound) {
			return nil, sharedErrs.NewUnauthorized(err)
		}
		return nil, sharedErrs.NewInternal(err)
	}
	return converter.UserSessionToProto(user, session), nil
}

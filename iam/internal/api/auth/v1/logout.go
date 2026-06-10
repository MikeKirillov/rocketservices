package v1

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if req.GetSessionUuid() == "" {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrEmptySessionID)
	}
	sessionUUID, err := uuid.Parse(req.GetSessionUuid())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrInvalidUUID)
	}

	if err := a.authService.Logout(ctx, sessionUUID); err != nil {
		return nil, err
	}
	return &authv1.LogoutResponse{}, nil
}

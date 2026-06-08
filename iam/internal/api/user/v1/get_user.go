package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/api/converter"
	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func (a *api) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	id, err := uuid.Parse(req.GetUserUuid())
	if err != nil {
		return nil, sharedErrs.NewInvalidArgument(errs.ErrInvalidUUID)
	}

	user, err := a.userService.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}

	return converter.UserToProto(user), nil
}

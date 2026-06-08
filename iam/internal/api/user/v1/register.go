package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/iam/internal/api/converter"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func (a *api) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	input := converter.RegisterToInput(req)

	userId, err := a.userService.Register(ctx, input)
	if err != nil {
		return nil, err
	}

	return &userv1.RegisterResponse{
		UserUuid: userId.String(),
	}, nil
}

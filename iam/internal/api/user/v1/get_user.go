package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func (a *api) GetUser(context.Context, *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetUser not implemented")
}

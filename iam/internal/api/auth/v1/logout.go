package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Logout(context.Context, *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Logout not implemented")
}

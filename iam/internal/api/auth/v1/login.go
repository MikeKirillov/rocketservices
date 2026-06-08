package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Login(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Login not implemented")
}

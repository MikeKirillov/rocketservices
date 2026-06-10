package v1

import (
	"context"

	"github.com/MikeKirillov/rocketservices/iam/internal/api/converter"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	input := converter.LoginToInput(req)

	sessionUUID, err := a.authService.Login(ctx, input)
	if err != nil {
		return nil, err
	}

	return &authv1.LoginResponse{
		SessionUuid: sessionUUID.String(),
	}, nil
}

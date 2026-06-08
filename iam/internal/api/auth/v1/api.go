package v1

import (
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

type api struct {
	authv1.UnimplementedAuthServiceServer

	authService AuthService
}

func NewAPI(authService AuthService) *api {
	return &api{authService: authService}
}

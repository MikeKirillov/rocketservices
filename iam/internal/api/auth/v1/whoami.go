package v1

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
)

func (a *api) Whoami(context.Context, *authv1.WhoamiRequest) (*authv1.WhoamiResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method Whoami not implemented")
}

// Whoami
// Ситуация						gRPC код			Описание
// Сессия валидна				OK					Возвращает session + user
// Пустой session_uuid			InvalidArgument		"session_uuid обязателен"
// Сессия не найдена / истекла	Unauthenticated		"сессия не найдена или истекла"

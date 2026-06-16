package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
)

type AuthService interface {
	Login(context.Context, input.LoginPassInput) (uuid.UUID, error)
	Whoami(context.Context, uuid.UUID) (model.User, model.Session, error)
	Logout(context.Context, uuid.UUID) error
}

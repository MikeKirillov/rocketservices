package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
)

type UserService interface {
	Register(context.Context, input.LoginPassInput) (uuid.UUID, error)
	GetUser(context.Context, uuid.UUID) (model.User, error)
}

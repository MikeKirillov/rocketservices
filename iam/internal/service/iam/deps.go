package iam

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
)

type UserRepository interface {
	InsertUser(context.Context, model.User) error
	GetUserByUUID(context.Context, uuid.UUID) (model.User, error)
	GetUserByLogin(context.Context, string) (model.User, error)
}

type SessionRepository interface {
	Set(context.Context, model.Session) error
	Delete(context.Context, uuid.UUID) error
}

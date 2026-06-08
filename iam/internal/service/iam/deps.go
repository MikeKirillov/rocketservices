package iam

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
)

type UserRepository interface {
	InsertUser(context.Context, model.User) error
	GetUser(context.Context, uuid.UUID) (model.User, error)
}

type SessionRepository interface{}

package iam

import (
	"context"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
)

type UserRepository interface {
	InsertUser(context.Context, model.User) error
	GetUser() error
}

type SessionRepository interface{}

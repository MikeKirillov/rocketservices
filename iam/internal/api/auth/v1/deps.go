package v1

import (
	"context"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/service/input"
)

// для обоих хендлеров используется интерфейс с одинаковым
// набором методов. по ТЗ у нас универсальный сервис для user и auth.
type AuthService interface {
	Register(context.Context, input.RegisterInput) (uuid.UUID, error)
	GetUser(context.Context, uuid.UUID) (model.User, error)
	Login(context.Context) error
	Whoami(context.Context) error
	Logout(context.Context) error
}

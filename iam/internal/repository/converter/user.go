package converter

import (
	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/repository/record"
)

func UserToRecord(user model.User) record.User {
	return record.User{
		Uuid:         uuid.New(),
		Login:        user.Login,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}
}

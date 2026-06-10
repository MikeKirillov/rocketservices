package converter

import (
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/repository/record"
)

func UserToRecord(user model.User) record.User {
	return record.User{
		Uuid:         user.UUID,
		Login:        user.Login,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}
}

func UserToModel(user record.User) model.User {
	return model.User{
		UUID:         user.Uuid,
		Login:        user.Login,
		CreatedAt:    user.CreatedAt,
		UpdatetAt:    user.UpdatedAt,
		PasswordHash: user.PasswordHash,
	}
}

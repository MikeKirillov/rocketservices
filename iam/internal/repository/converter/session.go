package converter

import (
	"time"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	redisview "github.com/MikeKirillov/rocketservices/iam/internal/repository/redis_view"
)

func SessionToView(s model.Session, u model.User) redisview.SessionRedisView {
	var userUpdated int64
	if u.UpdatedAt != nil {
		userUpdated = u.UpdatedAt.UnixNano()
	}

	return redisview.SessionRedisView{
		Uuid:             s.UUID.String(),
		UserUuid:         u.UUID.String(),
		UserLogin:        u.Login,
		UserPasswordHash: u.Password.Hash(),
		UserCreatedAt:    u.CreatedAt.UnixNano(),
		UserUpdatedAt:    userUpdated,
		CreatedAt:        s.CreatedAt.UnixNano(),
		ExpiresAt:        s.ExpiresAt.UnixNano(),
	}
}

func SessionToModel(s redisview.SessionRedisView) (model.Session, model.User) {
	user := model.User{
		UUID:      uuid.MustParse(s.UserUuid),
		Login:     s.UserLogin,
		Password:  model.RestorePassword(s.UserPasswordHash),
		CreatedAt: time.Unix(0, s.UserCreatedAt),
	}

	if s.UserUpdatedAt != 0 {
		user.UpdatedAt = new(time.Unix(0, s.UserUpdatedAt))
	}

	return model.Session{
		UUID:      uuid.MustParse(s.Uuid),
		CreatedAt: time.Unix(0, s.CreatedAt),
		ExpiresAt: time.Unix(0, s.ExpiresAt),
	}, user
}

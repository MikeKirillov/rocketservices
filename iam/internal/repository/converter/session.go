package converter

import (
	"time"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	redisview "github.com/MikeKirillov/rocketservices/iam/internal/repository/redis_view"
)

func SessionToView(s model.Session, expiration time.Duration) redisview.SessionRedisView {
	return redisview.SessionRedisView{
		Uuid:      s.UUID.String(),
		UserUuid:  s.UserUUID.String(),
		Login:     s.Login,
		CreatedAt: s.CreatedAt.UnixNano(),
		ExpiresAt: s.CreatedAt.Add(expiration).UnixNano(),
	}
}

func SessionToModel(s redisview.SessionRedisView) model.Session {
	return model.Session{
		UUID:      uuid.MustParse(s.Uuid),
		UserUUID:  uuid.MustParse(s.UserUuid),
		Login:     s.Login,
		CreatedAt: time.Unix(0, s.CreatedAt),
		ExpiresAt: time.Unix(0, s.ExpiresAt),
	}
}

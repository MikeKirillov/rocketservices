package converter

import (
	"time"

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

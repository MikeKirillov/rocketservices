package sessioncache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/repository/converter"
)

const cacheKeyPrefix = "session"

type repository struct {
	client redisClient
	ttl    time.Duration
}

func NewRepository(client redisClient, ttl time.Duration) *repository {
	return &repository{
		client: client,
		ttl:    ttl,
	}
}

func (r *repository) getChacheKey(uuid string) string {
	return fmt.Sprintf("%s:%s", cacheKeyPrefix, uuid)
}

// DEL идемпотентен — повторный Logout для несуществующей сессии не ошибка.

func (r *repository) Set(ctx context.Context, session model.Session) error {
	cacheKey := r.getChacheKey(session.UUID.String())

	err := r.client.HSet(ctx, cacheKey, new(converter.SessionToView(session, r.ttl))).Err()
	if err != nil {
		return err
	}

	return r.client.Expire(ctx, cacheKey, r.ttl).Err()
}

func (r *repository) Delete(ctx context.Context, sessionUUID uuid.UUID) error {
	cacheKey := r.getChacheKey(sessionUUID.String())
	return r.client.Del(ctx, cacheKey).Err()
}

package sessioncache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	errs "github.com/MikeKirillov/rocketservices/iam/internal/errors"
	"github.com/MikeKirillov/rocketservices/iam/internal/model"
	"github.com/MikeKirillov/rocketservices/iam/internal/repository/converter"
	redisview "github.com/MikeKirillov/rocketservices/iam/internal/repository/redis_view"
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

func (r *repository) Get(ctx context.Context, sessionUUID uuid.UUID) (model.Session, error) {
	cacheKey := r.getChacheKey(sessionUUID.String())

	var session redisview.SessionRedisView
	err := r.client.HGetAll(ctx, cacheKey).Scan(&session)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.Session{}, errs.ErrSessionNotFound
		}
		return model.Session{}, err
	}

	if session.Uuid == "" {
		return model.Session{}, errs.ErrSessionNotFound
	}

	return converter.SessionToModel(session), nil
}

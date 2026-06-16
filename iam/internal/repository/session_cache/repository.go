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
}

func NewRepository(client redisClient) *repository {
	return &repository{
		client: client,
	}
}

func (r *repository) getChacheKey(uuid string) string {
	return fmt.Sprintf("%s:%s", cacheKeyPrefix, uuid)
}

func (r *repository) Set(ctx context.Context, session model.Session, user model.User, ttl time.Duration) error {
	cacheKey := r.getChacheKey(session.UUID.String())

	err := r.client.HSet(ctx, cacheKey, new(converter.SessionToView(session, user))).Err()
	if err != nil {
		return err
	}

	return r.client.Expire(ctx, cacheKey, ttl).Err()
}

func (r *repository) Delete(ctx context.Context, sessionUUID uuid.UUID) error {
	cacheKey := r.getChacheKey(sessionUUID.String())
	return r.client.Del(ctx, cacheKey).Err()
}

func (r *repository) Get(ctx context.Context, sessionUUID uuid.UUID) (model.Session, model.User, error) {
	cacheKey := r.getChacheKey(sessionUUID.String())

	var session redisview.SessionRedisView
	err := r.client.HGetAll(ctx, cacheKey).Scan(&session)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.Session{}, model.User{}, errs.ErrSessionNotFound
		}
		return model.Session{}, model.User{}, err
	}

	if session.Uuid == "" {
		return model.Session{}, model.User{}, errs.ErrSessionNotFound
	}

	s, u := converter.SessionToModel(session)
	return s, u, nil
}

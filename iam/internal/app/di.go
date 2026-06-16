package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	authV1API "github.com/MikeKirillov/rocketservices/iam/internal/api/auth/v1"
	userV1API "github.com/MikeKirillov/rocketservices/iam/internal/api/user/v1"
	"github.com/MikeKirillov/rocketservices/iam/internal/config"
	sessionRepo "github.com/MikeKirillov/rocketservices/iam/internal/repository/session_cache"
	userRepo "github.com/MikeKirillov/rocketservices/iam/internal/repository/user"
	iamService "github.com/MikeKirillov/rocketservices/iam/internal/service/iam"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	platformRedis "github.com/MikeKirillov/rocketservices/platform/pkg/redis"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

type diContainer struct {
	// Инфраструктура
	pgPool      *pgxpool.Pool
	redisClient *redis.Client

	// Репозитории
	userRepo    iamService.UserRepository
	sessionRepo iamService.SessionRepository

	// Сервисы
	userSvc userV1API.UserService
	authSvc authV1API.AuthService

	// API-обработчики
	authV1Handler authv1.AuthServiceServer
	userV1Handler userv1.UserServiceServer
}

func (d *diContainer) PGPool(ctx context.Context) *pgxpool.Pool {
	if d.pgPool == nil {
		pool, err := pgxpool.New(ctx, config.AppConfig().PG.DSN())
		if err != nil {
			slog.Error("ошибка подключения к БД", "error", err)
			os.Exit(1)
		}

		err = pool.Ping(ctx)
		if err != nil {
			slog.Error("проверка соединения с БД", "error", err)
			os.Exit(1)
		}

		slog.Info("подключение к PostgreSQL установлено")

		closer.Add("PostgreSQL pool", func(_ context.Context) error {
			pool.Close()
			return nil
		})

		d.pgPool = pool
	}
	return d.pgPool
}

func (d *diContainer) UserRepository(ctx context.Context) iamService.UserRepository {
	if d.userRepo == nil {
		d.userRepo = userRepo.NewRepository(d.PGPool(ctx))
	}
	return d.userRepo
}

func (d *diContainer) RedisClient() *redis.Client {
	if d.redisClient == nil {
		rdb, err := platformRedis.NewClient(
			&redis.Options{
				Addr:         config.AppConfig().Redis.Address(),
				DialTimeout:  config.AppConfig().Redis.ConnTimeout(),
				ReadTimeout:  config.AppConfig().Redis.ConnTimeout(),
				WriteTimeout: config.AppConfig().Redis.ConnTimeout(),
			},
			slog.Default(),
		)
		if err != nil {
			slog.Error("не удалось создать Redis клиент", "error", err)
			os.Exit(1)
		}

		closer.Add("Redis", func(_ context.Context) error {
			return rdb.Close()
		})

		d.redisClient = rdb
	}
	return d.redisClient
}

func (d *diContainer) SessionRepository() iamService.SessionRepository {
	if d.sessionRepo == nil {
		d.sessionRepo = sessionRepo.NewRepository(d.RedisClient())
	}
	return d.sessionRepo
}

func (d *diContainer) AuthService(ctx context.Context) authV1API.AuthService {
	if d.authSvc == nil {
		d.authSvc = iamService.New(
			d.UserRepository(ctx),
			d.SessionRepository(),
			config.AppConfig().Redis.SessionTTL(),
		)
	}
	return d.authSvc
}

func (d *diContainer) UserService(ctx context.Context) userV1API.UserService {
	if d.userSvc == nil {
		d.userSvc = d.AuthService(ctx) // т.к. для обоих хендлеров используется интерфейс с одинаковым
		// набором методов. по ТЗ у нас универсальный сервис для user и auth
	}
	return d.userSvc
}

func (d *diContainer) AuthV1API(ctx context.Context) authv1.AuthServiceServer {
	if d.authV1Handler == nil {
		d.authV1Handler = authV1API.NewAPI(d.UserService(ctx))
	}
	return d.authV1Handler
}

func (d *diContainer) UserV1API(ctx context.Context) userv1.UserServiceServer {
	if d.userV1Handler == nil {
		d.userV1Handler = userV1API.NewAPI(d.UserService(ctx))
	}
	return d.userV1Handler
}

package app

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"

	authAPIV1 "github.com/MikeKirillov/rocketservices/iam/internal/api/auth/v1"
	userAPIV1 "github.com/MikeKirillov/rocketservices/iam/internal/api/user/v1"
	sessionRepository "github.com/MikeKirillov/rocketservices/iam/internal/repository/session_cache"
	userRepository "github.com/MikeKirillov/rocketservices/iam/internal/repository/user"
	iamService "github.com/MikeKirillov/rocketservices/iam/internal/service/iam"
	"github.com/MikeKirillov/rocketservices/platform/pkg/grpc/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
	userv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/user/v1"
)

func NewGRPCServer(
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	ttl time.Duration,
	minCost int,
) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	)
	registerServices(grpcServer, pool, redisClient, ttl)
	return grpcServer
}

func registerServices(
	grpcServer *grpc.Server,
	pool *pgxpool.Pool,
	redisClient *redis.Client,
	ttl time.Duration,
) {
	userRepo := userRepository.NewRepository(pool)
	sessionRepo := sessionRepository.NewRepository(redisClient, ttl)

	svc := iamService.New(userRepo, sessionRepo)

	authAPI := authAPIV1.NewAPI(svc)
	userAPI := userAPIV1.NewAPI(svc)

	authv1.RegisterAuthServiceServer(grpcServer, authAPI)
	userv1.RegisterUserServiceServer(grpcServer, userAPI)
}

package app

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	partAPIV1 "github.com/MikeKirillov/rocketservices/inventory/internal/api/part/v1"
	partRepository "github.com/MikeKirillov/rocketservices/inventory/internal/repository/part"
	partService "github.com/MikeKirillov/rocketservices/inventory/internal/service/part"
	interceptor "github.com/MikeKirillov/rocketservices/shared/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func RegisterServices(grpcServer *grpc.Server, pool *pgxpool.Pool) {
	repo := partRepository.NewRepository(pool)
	svc := partService.NewService(repo)
	api := partAPIV1.NewAPI(svc)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	}
}

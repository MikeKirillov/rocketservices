package app

import (
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	partAPIV1 "github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1"
	partService "github.com/MikeKirillov/rocketservices/inventory/internal/application/service"
	authClientV1 "github.com/MikeKirillov/rocketservices/inventory/internal/client/grpc/auth/v1"
	partComp "github.com/MikeKirillov/rocketservices/inventory/internal/domain/service"
	innerceptor "github.com/MikeKirillov/rocketservices/inventory/internal/interceptor"
	partRepository "github.com/MikeKirillov/rocketservices/inventory/internal/repository/part"
	"github.com/MikeKirillov/rocketservices/platform/pkg/grpc/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

func RegisterServices(
	grpcServer *grpc.Server,
	pool *pgxpool.Pool,
	txManager *manager.Manager,
) {
	repo := partRepository.NewRepository(pool)
	comp := partComp.NewCompatibilityChecker()
	svc := partService.NewService(repo, comp, txManager)
	api := partAPIV1.NewAPI(svc)
	inventoryv1.RegisterInventoryServiceServer(grpcServer, api)
}

func Interceptors(authGRPCClient authv1.AuthServiceClient) []grpc.ServerOption {
	authClient := authClientV1.New(authGRPCClient)
	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			innerceptor.ServerIncomingInterceptor(authClient),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	}
}

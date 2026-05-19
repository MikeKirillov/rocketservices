package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	orderV1API "github.com/MikeKirillov/rocketservices/order/internal/api/order/v1"
	inventoryClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/payment/v1"
	orderConfig "github.com/MikeKirillov/rocketservices/order/internal/config"
	orderRepository "github.com/MikeKirillov/rocketservices/order/internal/repository/order"
	orderService "github.com/MikeKirillov/rocketservices/order/internal/service/order"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

type diContainer struct {
	pgPool          *pgxpool.Pool
	txManager       *manager.Manager
	orderRepo       orderService.OrderRepository
	inventoryClient orderService.InventoryServiceClient
	paymentClient   orderService.PaymentServiceClient
	orderSvc        orderV1API.OrderService
	orderV1Handler  orderv1.Handler
}

func (d *diContainer) PGPool(ctx context.Context) *pgxpool.Pool {
	if d.pgPool == nil {
		pool, err := pgxpool.New(ctx, orderConfig.AppConfig().PG.DSN())
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

func (d *diContainer) TxManager(ctx context.Context) *manager.Manager {
	if d.txManager == nil {
		txManager, err := manager.New(trmpgx.NewDefaultFactory(d.PGPool(ctx)))
		if err != nil {
			slog.Error("создание transaction manager", "error", err)
			os.Exit(1)
		}
		d.txManager = txManager
	}
	return d.txManager
}

func (d *diContainer) OrderRepository(ctx context.Context) orderService.OrderRepository {
	if d.orderRepo == nil {
		d.orderRepo = orderRepository.NewRepository(
			d.PGPool(ctx),
			d.TxManager(ctx),
		)
	}
	return d.orderRepo
}

func (d *diContainer) InventoryClient(ctx context.Context) orderService.InventoryServiceClient {
	config := orderConfig.AppConfig().GRPC.InventoryGRPCConfig

	if d.inventoryClient == nil {
		conn, err := newGRPCClient(config.Address(), config.Time(), config.Timeout())
		if err != nil {
			slog.Error("не удалось подключиться к gRPC: InventoryService", "error", err)
			os.Exit(1)
		}

		closer.Add("gRPC соединение InventoryService", func(_ context.Context) error {
			if err = conn.Close(); err != nil {
				slog.Error("ошибка закрытия gRPC соединения", "error", err)
				return err
			}
			return nil
		})

		grpcClient := inventoryv1.NewInventoryServiceClient(conn)
		d.inventoryClient = inventoryClientV1.New(grpcClient)
	}
	slog.Info("установлено подключение к gRPC: InventoryService", "host", config.Host, "port", config.Port)

	return d.inventoryClient
}

func (d *diContainer) PaymentClient(ctx context.Context) orderService.PaymentServiceClient {
	config := orderConfig.AppConfig().GRPC.PaymentGRPCConfig

	if d.paymentClient == nil {
		conn, err := newGRPCClient(config.Address(), config.Time(), config.Timeout())
		if err != nil {
			slog.Error("не удалось подключиться к gRPC: PaymentService", "error", err)
			os.Exit(1)
		}

		closer.Add("gRPC соединение PaymentService", func(_ context.Context) error {
			if err = conn.Close(); err != nil {
				slog.Error("ошибка закрытия gRPC соединения", "error", err)
				return err
			}
			return nil
		})

		grpcClient := paymentv1.NewPaymentServiceClient(conn)
		d.paymentClient = paymentClientV1.New(grpcClient)
	}
	slog.Info("установлено подключение к gRPC: PaymentService", "host", config.Host, "port", config.Port)

	return d.paymentClient
}

func newGRPCClient(address string, time, timeout time.Duration) (conn *grpc.ClientConn, err error) {
	return grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                time,    // Интервал ping'ов
			Timeout:             timeout, // Таймаут ожидания pong
			PermitWithoutStream: true,    // Пинговать даже без активных RPC
		}),
	)
}

func (d *diContainer) OrderService(ctx context.Context) orderV1API.OrderService {
	if d.orderSvc == nil {
		d.orderSvc = orderService.NewService(
			d.OrderRepository(ctx),
			d.InventoryClient(ctx),
			d.PaymentClient(ctx),
			d.TxManager(ctx),
		)
	}
	return d.orderSvc
}

func (d *diContainer) OrderV1API(ctx context.Context) orderv1.Handler {
	if d.orderV1Handler == nil {
		d.orderV1Handler = orderV1API.NewOrderHandler(d.OrderService(ctx))
	}
	return d.orderV1Handler
}

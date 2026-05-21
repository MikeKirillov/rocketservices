package app

import (
	"context"
	"log/slog"
	"os"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"

	invV1API "github.com/MikeKirillov/rocketservices/inventory/internal/api/inventory/v1"
	invService "github.com/MikeKirillov/rocketservices/inventory/internal/application/service"
	"github.com/MikeKirillov/rocketservices/inventory/internal/config"
	domain "github.com/MikeKirillov/rocketservices/inventory/internal/domain/service"
	invRepository "github.com/MikeKirillov/rocketservices/inventory/internal/repository/part"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

type diContainer struct {
	pgPool       *pgxpool.Pool
	txManager    *manager.Manager
	invRepo      invService.PartRepository
	compChecker  invService.CompatibilityChecker
	invSvc       invV1API.PartService
	invV1Handler inventoryv1.InventoryServiceServer
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

func (d *diContainer) InventoryRepository(ctx context.Context) invService.PartRepository {
	if d.invRepo == nil {
		d.invRepo = invRepository.NewRepository(d.PGPool(ctx))
	}
	return d.invRepo
}

func (d *diContainer) CompatibilityChecker() invService.CompatibilityChecker {
	if d.compChecker == nil {
		d.compChecker = domain.NewCompatibilityChecker()
	}
	return d.compChecker
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

func (d *diContainer) InventoryService(ctx context.Context) invV1API.PartService {
	if d.invSvc == nil {
		d.invSvc = invService.NewService(
			d.InventoryRepository(ctx),
			d.CompatibilityChecker(),
			d.TxManager(ctx),
		)
	}
	return d.invSvc
}

func (d *diContainer) InventoryV1API(ctx context.Context) inventoryv1.InventoryServiceServer {
	if d.invV1Handler == nil {
		d.invV1Handler = invV1API.NewAPI(d.InventoryService(ctx))
	}
	return d.invV1Handler
}

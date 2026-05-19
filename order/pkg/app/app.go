package app

import (
	"log/slog"
	"net/http"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"

	orderAPIV1 "github.com/MikeKirillov/rocketservices/order/internal/api/order/v1"
	inventoryClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/payment/v1"
	"github.com/MikeKirillov/rocketservices/order/internal/middleware"
	orderRepo "github.com/MikeKirillov/rocketservices/order/internal/repository/order"
	orderService "github.com/MikeKirillov/rocketservices/order/internal/service/order"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func NewHTTPHandler(
	pool *pgxpool.Pool,
	txManager *manager.Manager,
	inventoryGRPCClient inventoryv1.InventoryServiceClient,
	paymentGRPCClient paymentv1.PaymentServiceClient,
) (http.Handler, error) {
	repo := orderRepo.NewRepository(pool, txManager)
	inventoryClient := inventoryClientV1.New(inventoryGRPCClient)
	paymentClient := paymentClientV1.New(paymentGRPCClient)
	service := orderService.NewService(repo, inventoryClient, paymentClient, txManager)
	orderHandler := orderAPIV1.NewOrderHandler(service)

	// Создать OpenAPI сервер на основе обработчика
	orderServer, err := orderv1.NewServer(orderHandler)
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", orderServer)

	var handler http.Handler = mux

	// Применяем кастомные middleware
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)

	return handler, nil
}

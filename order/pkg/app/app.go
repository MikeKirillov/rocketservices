package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/IBM/sarama"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"

	orderAPIV1 "github.com/MikeKirillov/rocketservices/order/internal/api/order/v1"
	authClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/auth/v1"
	inventoryClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/inventory/v1"
	paymentClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/payment/v1"
	"github.com/MikeKirillov/rocketservices/order/internal/middleware"
	"github.com/MikeKirillov/rocketservices/order/internal/model"
	producer "github.com/MikeKirillov/rocketservices/order/internal/producer/order"
	orderRepo "github.com/MikeKirillov/rocketservices/order/internal/repository/order"
	orderService "github.com/MikeKirillov/rocketservices/order/internal/service/order"
	wrappedKafkaProducer "github.com/MikeKirillov/rocketservices/platform/pkg/kafka/producer"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
	authv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/auth/v1"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

func NewHTTPHandler(
	pool *pgxpool.Pool,
	txManager *manager.Manager,
	inventoryGRPCClient inventoryv1.InventoryServiceClient,
	paymentGRPCClient paymentv1.PaymentServiceClient,
	authGRPCClient authv1.AuthServiceClient,
) (http.Handler, error) {
	repo := orderRepo.NewRepository(pool, txManager)
	inventoryClient := inventoryClientV1.New(inventoryGRPCClient)
	paymentClient := paymentClientV1.New(paymentGRPCClient)
	authClient := authClientV1.New(authGRPCClient)

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, cfg)
	if err != nil {
		slog.Error("не удалось создать sync producer", "error", err)
		return nil, err
	}
	orderPaidProducer := wrappedKafkaProducer.NewProducer(syncProducer, "order.paid")
	producerService := producer.NewService(orderPaidProducer)

	service := orderService.NewService(repo, inventoryClient, paymentClient, authClient, txManager, producerService)
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

type OrderProducerService interface {
	ProduceOrderPaid(ctx context.Context, order model.Order) error
}

func NewHTTPHandlerWithProducer(
	pool *pgxpool.Pool,
	txManager *manager.Manager,
	inventoryGRPCClient inventoryv1.InventoryServiceClient,
	paymentGRPCClient paymentv1.PaymentServiceClient,
	realOrderProducer OrderProducerService,
	authGRPCClient authv1.AuthServiceClient,
) (http.Handler, error) {
	repo := orderRepo.NewRepository(pool, txManager)
	inventoryClient := inventoryClientV1.New(inventoryGRPCClient)
	paymentClient := paymentClientV1.New(paymentGRPCClient)
	authClient := authClientV1.New(authGRPCClient)

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V4_0_0_0
	cfg.Producer.Return.Successes = true

	service := orderService.NewService(repo, inventoryClient, paymentClient, authClient, txManager, realOrderProducer)
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

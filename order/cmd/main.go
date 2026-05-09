package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/MikeKirillov/rocketservices/order/pkg/app"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

const (
	inventoryServiceAddress = "localhost:50051"
	paymentServiceAddress   = "localhost:50052"
	httpPort                = "8080"

	// HTTP-Server таймауты.
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
	middlewareTimeout = 10 * time.Second

	// gRPC клиент keepalive параметры.
	grpcKeepaliveTime    = 5 * time.Minute // Интервал ping'ов для обнаружения мёртвого сервера
	grpcKeepaliveTimeout = 1 * time.Second // Таймаут ожидания pong
)

func main() {
	// При получении сигнала SIGINT/SIGTERM сервер должен:
	// 1. Перестать принимать новые соединения
	// 2. Дождаться завершения текущих запросов (с таймаутом)
	// 3. Закрыть gRPC соединения
	// 4. Корректно завершить работу
	// Подсказка: используйте signal.Notify и httpServer.Shutdown(ctx)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, txManager, err := createPgConnection(ctx)
	if err != nil {
		cancel()
		os.Exit(1) //nolint:gocritic // cancel() выполняется непосредственно перед os.Exit(1)
	}

	// Настроить gRPC клиент с параметрами keepalive
	// Подумайте, какие параметры стоит задать для gRPC клиента
	// См. examples/week_1/GRPC_CONNECTIONS.md

	// Создать gRPC соединение с InventoryService
	inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcKeepaliveTime,    // Интервал ping'ов
			Timeout:             grpcKeepaliveTimeout, // Таймаут ожидания pong
			PermitWithoutStream: true,                 // Пинговать даже без активных RPC
		}))
	if err != nil {
		slog.Error("не удалось подключиться к gRPC: InventoryService", "error", err)
		os.Exit(1)
	}
	inventoryGRPCClient := inventoryv1.NewInventoryServiceClient(inventoryConn)

	// Создать gRPC клиент PaymentService
	paymentConn, err := grpc.NewClient(paymentServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcKeepaliveTime,    // Интервал ping'ов
			Timeout:             grpcKeepaliveTimeout, // Таймаут ожидания pong
			PermitWithoutStream: true,                 // Пинговать даже без активных RPC
		}))
	if err != nil {
		slog.Error("не удалось подключиться к gRPC: PaymentService", "error", err)
		os.Exit(1)
	}
	paymentGRPCClient := paymentv1.NewPaymentServiceClient(paymentConn)

	defer func() {
		if err = inventoryConn.Close(); err != nil {
			slog.Error("ошибка закрытия gRPC соединения", "error", err)
		}
		if err = paymentConn.Close(); err != nil {
			slog.Error("ошибка закрытия gRPC соединения", "error", err)
		}
	}()

	// Создаём хранилище, сервис и обработчик
	handler, err := app.NewHTTPHandler(pool, txManager, inventoryGRPCClient, paymentGRPCClient)
	if err != nil {
		slog.Error("ошибка создания обработчика HTTP-запросов", "error", err)
		return
	}
	handler = http.TimeoutHandler(handler, middlewareTimeout, "таймаут при обработке запроса к серверу")

	// Настроить HTTP сервер с таймаутами
	// Подумайте, какие таймауты стоит задать для production-ready сервера
	// См. examples/week_1/HTTP_SERVER.md
	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атаки
		ReadTimeout:       readTimeout,       // Лимит на чтение всего запроса
		WriteTimeout:      writeTimeout,      // Лимит на запись ответа
		IdleTimeout:       idleTimeout,       // Таймаут keep-alive соединений
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		slog.Info("запуск сервера OrderService", "port", httpPort)

		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ошибка запуска сервера", "error", err)
			cancel()
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	slog.Info("завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("ошибка при остановке сервера", "error", err)
	}
	slog.Info("сервер остановлен")
}

func createPgConnection(ctx context.Context) (*pgxpool.Pool, *manager.Manager, error) {
	// Загружаем переменные окружения из .env
	err := godotenv.Load("order.env")
	if err != nil {
		slog.Error("ошибка загрузки переменных окружения из inventory.env", "error", err)
		return nil, nil, err
	}

	// Подключаемся к PostgreSQL
	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		slog.Error("переменная окружения DB_URI не установлена")
		return nil, nil, err
	}

	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		slog.Error("ошибка подключения к БД", "error", err)
		return nil, nil, err
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		slog.Error("проверка соединения с БД", "error", err)
		return nil, nil, err
	}

	slog.Info("подключение к PostgreSQL установлено")

	// Создаём Transaction Manager для pgx
	txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
	if err != nil {
		slog.Error("создание transaction manager", "error", err)
		return nil, nil, err
	}

	return pool, txManager, nil
}

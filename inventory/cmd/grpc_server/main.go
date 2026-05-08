package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	invApp "github.com/MikeKirillov/rocketservices/inventory/pkg/app"
	"github.com/MikeKirillov/rocketservices/shared/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

const (
	grpcAddress = "localhost:50051"

	// gRPC keepalive параметры.
	grpcMaxConnectionIdle     = 15 * time.Minute // Закрыть idle-соединения (нет активных RPC)
	grpcMaxConnectionAge      = 30 * time.Minute // Принудительная ротация для балансировки
	grpcMaxConnectionAgeGrace = 5 * time.Second  // Время на завершение активных RPC
	grpcKeepaliveTime         = 5 * time.Minute  // Интервал ping'ов для обнаружения мёртвых соединений
	grpcKeepaliveTimeout      = 1 * time.Second  // Таймаут ожидания pong
	grpcMinPingInterval       = 5 * time.Minute  // Минимальный интервал ping'ов от клиента (защита от DoS)
)

func main() {
	// При получении сигнала SIGINT/SIGTERM сервер должен:
	// 1. Перестать принимать новые соединения
	// 2. Дождаться завершения текущих запросов
	// 3. Корректно завершить работу
	// Подсказка: используйте signal.Notify и grpcServer.GracefulStop()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Загружаем переменные окружения из .env
	err := godotenv.Load("inventory.env")
	if err != nil {
		slog.Error("ошибка загрузки переменных окружения из inventory.env", "error", err)
		cancel()
		os.Exit(1)
	}

	// Подключаемся к PostgreSQL
	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		slog.Error("переменная окружения DB_URI не установлена")
		cancel()
		os.Exit(1)
	}

	pool, err := pgxpool.New(ctx, dbURI)
	if err != nil {
		slog.Error("ошибка подключения к БД", "error", err)
		cancel()
		os.Exit(1)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		slog.Error("проверка соединения с БД", "error", err)
		pool.Close()
		cancel()
		os.Exit(1)
	}

	slog.Info("подключение к PostgreSQL установлено")

	//nolint:noctx // Контекст здесь не нужен: GracefulStop() сам закроет listener и прервёт Accept()
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		pool.Close()
		cancel()
		os.Exit(1)
	}

	// Настроить gRPC сервер с параметрами keepalive
	// Подумайте, какие параметры стоит задать для production-ready сервера
	// См. examples/week_1/GRPC_CONNECTIONS.md
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     grpcMaxConnectionIdle,
			MaxConnectionAge:      grpcMaxConnectionAge,
			MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
			Time:                  grpcKeepaliveTime,
			Timeout:               grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcMinPingInterval,
			PermitWithoutStream: true, // Разрешить "тёплые" соединения без активных RPC
		}),
		// Интерцепторы: recovery (перехват паник) + логирование запросов
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	)

	// Регистрируем наш сервис
	invApp.RegisterServices(grpcServer, pool)

	// Включаем reflection для postman/grpcurl
	reflection.Register(grpcServer)

	go func() {
		slog.Info("gRPC сервер InventoryService запущен", "адрес", grpcAddress)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("ошибка запуска сервера InventoryService", "error", err)
			cancel()
		}
	}()

	// Реализовать graceful shutdown
	<-ctx.Done()
	slog.Info("остановка gRPC сервера...")
	grpcServer.GracefulStop()
	slog.Info("сервер остановлен")
}

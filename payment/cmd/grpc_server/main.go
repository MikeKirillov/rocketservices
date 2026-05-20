package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	payApp "github.com/MikeKirillov/rocketservices/payment/pkg/app"
	"github.com/MikeKirillov/rocketservices/shared/interceptor"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
)

const (
	grpcAddress = "localhost:50052"

	// gRPC keepalive параметры.
	grpcMaxConnectionIdle     = 15 * time.Minute // Закрыть idle-соединения (нет активных RPC)
	grpcMaxConnectionAge      = 30 * time.Minute // Принудительная ротация для балансировки
	grpcMaxConnectionAgeGrace = 5 * time.Second  // Время на завершение активных RPC
	grpcKeepaliveTime         = 5 * time.Minute  // Интервал ping'ов для обнаружения мёртвых соединений
	grpcKeepaliveTimeout      = 1 * time.Second  // Таймаут ожидания pong
	grpcMinPingInterval       = 5 * time.Minute  // Минимальный интервал ping'ов от клиента (защита от DoS)
)

func main() {
	// Реализовать graceful shutdown
	// При получении сигнала SIGINT/SIGTERM сервер должен:
	// 1. Перестать принимать новые соединения
	// 2. Дождаться завершения текущих запросов
	// 3. Корректно завершить работу
	// Подсказка: используйте signal.Notify и grpcServer.GracefulStop()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//nolint:noctx // Контекст здесь не нужен: GracefulStop() сам закроет listener и прервёт Accept()
	lis, err := net.Listen("tcp", grpcAddress)
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		cancel()
		os.Exit(1) //nolint:gocritic // cancel() выполняется непосредственно перед os.Exit(1)
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
	payApp.RegisterServices(grpcServer)

	// Включаем reflection для postman/grpcurl
	reflection.Register(grpcServer)

	go func() {
		slog.Info("gRPC сервер PaymentService запущен", "адрес", grpcAddress)
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("ошибка запуска сервера PaymentService", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	slog.Info("остановка gRPC сервера...")
	grpcServer.GracefulStop()
	slog.Info("сервер остановлен")
}

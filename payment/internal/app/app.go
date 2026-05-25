package app

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/MikeKirillov/rocketservices/payment/internal/config"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	"github.com/MikeKirillov/rocketservices/platform/pkg/grpc/health"
	"github.com/MikeKirillov/rocketservices/platform/pkg/interceptor"
	"github.com/MikeKirillov/rocketservices/platform/pkg/logger"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	paymentv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/payment/v1"
)

// App — корневая структура приложения, управляющая жизненным циклом всех компонентов.
type App struct {
	diContainer *diContainer
	grpcServer  *grpc.Server
	listener    net.Listener
	configPath  string
}

// Run управляет жизненным циклом приложения: запускает серверы, обрабатывает сигналы ОС
// и выполняет graceful shutdown.
func (a *App) Run() error {
	// Реализовать graceful shutdown
	// При получении сигнала SIGINT/SIGTERM сервер должен:
	// 1. Перестать принимать новые соединения
	// 2. Дождаться завершения текущих запросов
	// 3. Корректно завершить работу
	// Подсказка: используйте signal.Notify и grpcServer.GracefulStop()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a.startGracefulShutdown(ctx, cancel)
	return a.runGRPCServer()
}

// startGracefulShutdown запускает горутину, которая ожидает сигнал завершения (SIGINT/SIGTERM)
// и корректно останавливает все компоненты приложения через closer.
//
// Порядок работы:
//  1. Ждём сигнал ОС — signal.NotifyContext перехватывает его и отменяет ctx
//  2. cancel() снимает перехват сигналов (вызывает signal.Stop внутри)
//     Контекст к этому моменту уже отменён. Это нужно, чтобы повторный Ctrl+C
//     не перехватывался, а завершил процесс принудительно (поведение ОС по умолчанию)
//  3. Создаём отдельный shutdownCtx от context.Background() с таймаутом — он не связан
//     с уже отменённым ctx, чтобы closer'ы имели гарантированное время на завершение.
func (a *App) startGracefulShutdown(ctx context.Context, cancel context.CancelFunc) {
	go func() { //nolint:gosec // G118: ctx уже отменён, context.Background нужен для graceful shutdown
		<-ctx.Done()
		cancel()
		slog.Info("получен сигнал завершения, начинаем graceful shutdown")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.AppConfig().GRPC.ShutdownTimeout())
		defer shutdownCancel()

		if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
			slog.Error("ошибка при завершении работы", "error", closeErr)
		}
		slog.Info("сервер остановлен")
	}()
}

// runGRPCServer запускает gRPC-сервер и блокирует до его остановки.
func (a *App) runGRPCServer() error {
	slog.Info("🚀 gRPC-сервер запущен", "address", config.AppConfig().GRPC.Address())

	return a.grpcServer.Serve(a.listener)
}

// New создает и инициализирует приложение.
func New(ctx context.Context, configPath string) *App {
	a := &App{configPath: configPath}
	a.initDeps(ctx)
	return a
}

// initDeps последовательно инициализирует все зависимости приложения.
func (a *App) initDeps(ctx context.Context) {
	inits := []func(context.Context){
		a.initConfig,
		a.initDi,
		a.initLogger,
		a.initListener,
		a.initGRPCServer,
	}

	for _, f := range inits {
		f(ctx)
	}
}

// initConfig загружает конфигурацию: сначала YAML, затем env-переменные поверх.
func (a *App) initConfig(ctx context.Context) {
	cfg, err := config.Load(a.configPath)
	if err != nil {
		slog.Error("ошибка загрузки конфигурации", "error", err)
		os.Exit(1)
	}

	slog.Info(
		"конфигурация загружена",
		"config_path", a.configPath,
		"grpc_address", cfg.GRPC.Address(),
	)
}

// initDI создаёт DI-контейнер.
func (a *App) initDi(ctx context.Context) {
	a.diContainer = &diContainer{}
}

// initLogger настраивает глобальный slog с уровнем из конфига.
func (a *App) initLogger(ctx context.Context) {
	logger.Init(config.AppConfig().Logger.Level)
}

// initListener создаёт TCP-листенер для gRPC-сервера.
func (a *App) initListener(ctx context.Context) {
	lis, err := net.Listen("tcp", config.AppConfig().GRPC.Address()) //nolint:noctx // Контекст здесь не нужен: GracefulStop() сам закроет listener и прервёт Accept()
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		os.Exit(1)
	}
	a.listener = lis
}

// initGRPCServer создаёт и настраивает gRPC-сервер, регистрирует обработчики.
func (a *App) initGRPCServer(ctx context.Context) {
	// Настроить gRPC сервер с параметрами keepalive
	// Подумайте, какие параметры стоит задать для production-ready сервера
	// См. examples/week_1/GRPC_CONNECTIONS.md
	a.grpcServer = grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     config.AppConfig().GRPC.MaxConnectionIdle(),
			MaxConnectionAge:      config.AppConfig().GRPC.MaxConnectionAge(),
			MaxConnectionAgeGrace: config.AppConfig().GRPC.MaxConnectionAgeGrace(),
			Time:                  config.AppConfig().GRPC.Time(),
			Timeout:               config.AppConfig().GRPC.Timeout(),
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.AppConfig().GRPC.MinPingInterval(),
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
	api := a.diContainer.PaymentV1API(ctx)

	closer.Add("gRPC server", func(ctx context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	// Включаем reflection для postman/grpcurl
	reflection.Register(a.grpcServer)

	// Регистрируем health service для проверки работоспособности
	health.RegisterService(a.grpcServer)

	paymentv1.RegisterPaymentServiceServer(a.grpcServer, api)
}

package app

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/MikeKirillov/rocketservices/inventory/internal/config"
	innerceptor "github.com/MikeKirillov/rocketservices/inventory/internal/interceptor"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	"github.com/MikeKirillov/rocketservices/platform/pkg/grpc/interceptor"
	"github.com/MikeKirillov/rocketservices/platform/pkg/logger"
	sharedErrs "github.com/MikeKirillov/rocketservices/shared/pkg/errors"
	inventoryv1 "github.com/MikeKirillov/rocketservices/shared/pkg/proto/inventory/v1"
)

type App struct {
	diContainer *diContainer
	grpcServer  *grpc.Server
	listener    net.Listener
	configPath  string
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a.startGracefulShutdown(ctx, cancel)
	return a.runGRPCServer()
}

func (a *App) startGracefulShutdown(ctx context.Context, cancel context.CancelFunc) {
	go func() { //nolint:gosec // G118: ctx уже отменён, context.Background нужен для graceful shutdown
		<-ctx.Done()
		cancel()
		slog.Info("получен сигнал завершения, начинаем graceful shutdown")
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			config.AppConfig().GRPC.ShutdownTimeout(),
		)
		defer shutdownCancel()

		if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
			slog.Error("ошибка при завершении работы", "error", closeErr)
		}
		slog.Info("сервер остановлен")
	}()
}

func (a *App) runGRPCServer() error {
	slog.Info("🚀 gRPC-сервер запущен", "address", config.AppConfig().GRPC.Address())

	return a.grpcServer.Serve(a.listener)
}

func New(ctx context.Context, configPath string) *App {
	a := &App{configPath: configPath}
	a.initDeps(ctx)
	return a
}

func (a *App) initDeps(ctx context.Context) {
	inits := []func(context.Context){
		a.initConfig,
		a.initDi,
		a.initLogger,
		a.initListener,
		a.initAuthClient,
		a.initGRPCServer,
	}

	for _, f := range inits {
		f(ctx)
	}
}

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
		"pg_host", cfg.PG.Host,
	)
}

func (a *App) initDi(ctx context.Context) {
	a.diContainer = &diContainer{}
}

func (a *App) initLogger(ctx context.Context) {
	logger.Init(logger.Config{
		Level:             config.AppConfig().Logger.Level,
		ServiceName:       config.AppConfig().Otel.ServiceName,
		Environment:       "development",
		EnableOTLP:        true,
		CollectorEndpoint: config.AppConfig().Otel.Endpoint,
	})
	closer.Add("логгер", func(_ context.Context) error {
		return logger.Close()
	})
}

func (a *App) initListener(ctx context.Context) {
	lis, err := net.Listen("tcp", config.AppConfig().GRPC.Address()) //nolint:noctx // Контекст здесь не нужен: GracefulStop() сам закроет listener и прервёт Accept()
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		os.Exit(1)
	}
	a.listener = lis
}

func (a *App) initAuthClient(_ context.Context) {
	a.diContainer.AuthClient()
}

func (a *App) initGRPCServer(ctx context.Context) {
	a.grpcServer = grpc.NewServer(
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
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.LoggerInterceptor(),
			innerceptor.ServerIncomingInterceptor(a.diContainer.authClient),
			sharedErrs.UnaryErrorInterceptor(slog.Default()),
		),
	)
	api := a.diContainer.InventoryV1API(ctx)

	closer.Add("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	reflection.Register(a.grpcServer)
	inventoryv1.RegisterInventoryServiceServer(a.grpcServer, api)
}

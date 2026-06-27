package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MikeKirillov/rocketservices/assembly/internal/config"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	"github.com/MikeKirillov/rocketservices/platform/pkg/logger"
)

const shutdownTimeout = 5 * time.Second

type App struct {
	diContainer *diContainer
	configPath  string
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a.startGracefulShutdown(ctx, cancel)
	return a.runConsumer(ctx)
}

func (a *App) startGracefulShutdown(ctx context.Context, cancel context.CancelFunc) {
	go func() { //nolint:gosec // G118: ctx уже отменён, context.Background нужен для graceful shutdown
		<-ctx.Done()
		cancel()
		slog.Info("получен сигнал завершения, начинаем graceful shutdown")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer shutdownCancel()

		if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
			slog.Error("ошибка при завершении работы", "error", closeErr)
		}
		slog.Info("сервер остановлен")
	}()
}

// runConsumer запускает Kafka-потребитель OrderPaid.
func (a *App) runConsumer(ctx context.Context) error {
	slog.Info("Kafka-потребитель OrderPaid запущен")

	return a.diContainer.OrderConsumerService().RunConsumer(ctx)
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

	slog.InfoContext(ctx,
		"конфигурация загружена",
		"config_path", a.configPath,
		"kafka_brokers", cfg.Kafka.Brokers,
	)
}

func (a *App) initDi(ctx context.Context) {
	a.diContainer = &diContainer{}
}

func (a *App) initLogger(ctx context.Context) {
	logger.Init(logger.Config{
		Level:             config.AppConfig().Logger.Level,
		ServiceName:       config.AppConfig().Otel.ServiceName,
		Environment:       config.AppConfig().Otel.Environment,
		EnableOTLP:        config.AppConfig().Otel.EnableOTLP,
		CollectorEndpoint: config.AppConfig().Otel.Endpoint,
	})
	closer.Add("logger", func(_ context.Context) error {
		return logger.Close()
	})
}

package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MikeKirillov/rocketservices/order/internal/config"
	"github.com/MikeKirillov/rocketservices/order/internal/middleware"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	"github.com/MikeKirillov/rocketservices/platform/pkg/logger"
	orderv1 "github.com/MikeKirillov/rocketservices/shared/pkg/openapi/order/v1"
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
	configPath  string
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a.startGracefulShutdown(ctx)

	errChan := make(chan error, 2)

	go func() {
		if err := a.runConsumer(ctx); err != nil {
			errChan <- fmt.Errorf("потребитель упал: %w", err)
		}
	}()
	go func() {
		if err := a.runHTTPServer(); err != nil {
			errChan <- fmt.Errorf("http-сервер упал: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (a *App) startGracefulShutdown(ctx context.Context) {
	go func() { //nolint:gosec // G118: ctx уже отменён, context.Background нужен для graceful shutdown
		<-ctx.Done()
		slog.Info("завершение работы сервера...")

		// Создаем контекст с таймаутом для остановки сервера
		shutdownCtx, shutdownCancel := context.WithTimeout(
			context.Background(),
			config.AppConfig().HTTP.TimeoutShutdown(),
		)
		defer shutdownCancel()

		if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
			slog.Error("ошибка при завершении работы", "error", closeErr)
		}

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("ошибка при остановке сервера", "error", err)
		}
		slog.Info("сервер остановлен")
	}()
}

func (a *App) runHTTPServer() error {
	slog.Info("запуск сервера OrderService", "port", config.AppConfig().HTTP.Address())

	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("ошибка запуска сервера", "error", err)
		return err
	}

	return nil
}

func (a *App) runConsumer(ctx context.Context) error {
	slog.Info("Kafka-потребитель AssemblyShipAssembled запущен")

	return a.diContainer.AssemblyConsumerService(ctx).RunConsumer(ctx)
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
		a.initHTTPServer,
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
		"http_address", cfg.HTTP.Address(),
		"pg_host", cfg.PG.Host,
	)
}

func (a *App) initDi(ctx context.Context) {
	a.diContainer = &diContainer{}
}

func (a *App) initLogger(ctx context.Context) {
	logger.Init(config.AppConfig().Logger.Level)
}

func (a *App) initHTTPServer(ctx context.Context) {
	orderHandler := a.diContainer.OrderV1API(ctx)

	// Создать OpenAPI сервер на основе обработчика
	orderServer, err := orderv1.NewServer(orderHandler)
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", orderServer)

	var handler http.Handler = mux

	// Применяем кастомные middleware
	handler = middleware.Logger(handler)
	handler = middleware.Recoverer(handler)
	handler = http.TimeoutHandler(
		handler,
		config.AppConfig().HTTP.TimeoutMiddlware(),
		"таймаут при обработке запроса к серверу",
	)

	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Address(),
		Handler:           handler,
		ReadHeaderTimeout: config.AppConfig().HTTP.TimeoutReadHeader(), // Защита от Slowloris атаки
		ReadTimeout:       config.AppConfig().HTTP.TimeoutRead(),       // Лимит на чтение всего запроса
		WriteTimeout:      config.AppConfig().HTTP.TimeoutWrite(),      // Лимит на запись ответа
		IdleTimeout:       config.AppConfig().HTTP.TimeoutIdle(),       // Таймаут keep-alive соединений
	}
}

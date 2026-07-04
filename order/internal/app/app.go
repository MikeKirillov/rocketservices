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

	"github.com/go-redis/redis_rate/v10"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/MikeKirillov/rocketservices/order/internal/config"
	"github.com/MikeKirillov/rocketservices/order/internal/middleware"
	"github.com/MikeKirillov/rocketservices/platform/pkg/closer"
	"github.com/MikeKirillov/rocketservices/platform/pkg/logger"
	"github.com/MikeKirillov/rocketservices/platform/pkg/metrics"
	"github.com/MikeKirillov/rocketservices/platform/pkg/ratelimit"
	"github.com/MikeKirillov/rocketservices/platform/pkg/tracing"
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
		a.initMetrics,
		a.initTracer,
		a.initAuthClient,
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

func (a *App) initMetrics(_ context.Context) {
	metrics.Init(
		config.AppConfig().Otel.ServiceName,
		config.AppConfig().Otel.Endpoint,
	)
	closer.Add("metrics provider", func(_ context.Context) error {
		return metrics.Close()
	})
}

func (a *App) initTracer(ctx context.Context) {
	shutdownTracer, err := tracing.InitTracer(
		ctx,
		tracing.Config{
			CollectorEndpoint: config.AppConfig().Otel.Endpoint,
			ServiceName:       config.AppConfig().Otel.ServiceName,
			Environment:       config.AppConfig().Otel.Environment,
			ServiceVersion:    config.AppConfig().Otel.ServiceVersion,
			SamplingRatio:     config.AppConfig().Otel.SamplingRatio,
		},
	)
	if err != nil {
		slog.Error("не удалось инициализировать трейсинг", "error", err)
		os.Exit(1)
	}
	closer.Add("tracing", func(ctx context.Context) error {
		return shutdownTracer(ctx)
	})
}

func (a *App) initAuthClient(_ context.Context) {
	a.diContainer.AuthClient()
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
	handler = middleware.Auth(handler, a.diContainer.authClient)
	handler = http.TimeoutHandler(
		handler,
		config.AppConfig().HTTP.TimeoutMiddlware(),
		"таймаут при обработке запроса к серверу",
	)
	// middleware для передачи телеметрии
	handler = otelhttp.NewHandler(handler, config.AppConfig().Otel.ServiceName)
	// middleware для подключения rate limit
	handler = ratelimit.Middleware(
		redis_rate.NewLimiter(a.diContainer.RedisClient()), // используем Redis для хранения данных о лимитах
		config.AppConfig().Redis.Rate,
		config.AppConfig().Redis.Burst,
		handler,
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

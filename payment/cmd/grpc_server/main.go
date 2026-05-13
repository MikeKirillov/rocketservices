package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/MikeKirillov/rocketservices/payment/internal/app"
	"github.com/MikeKirillov/rocketservices/payment/internal/config"
)

func main() {
	ctx := context.Background()

	// Определяем путь к конфиг-файлу: флаг -config > env CONFIG_PATH > config.local.yaml
	configPath := config.ResolveConfigPath()

	a := app.New(ctx, configPath)

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
		os.Exit(1)
	}
}

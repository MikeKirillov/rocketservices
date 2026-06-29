package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/MikeKirillov/rocketservices/order/internal/app"
	"github.com/MikeKirillov/rocketservices/order/internal/config"
)

func main() {
	ctx := context.Background()

	// Загружаем переменные окружения из .env
	if err := godotenv.Load("../order.env"); err != nil {
		slog.Error("ошибка загрузки переменных окружения из order.env", "error", err)
		// os.Exit(1)
	}

	configPath := config.ResolveConfigPath()
	a := app.New(ctx, configPath)

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
		os.Exit(1)
	}
}

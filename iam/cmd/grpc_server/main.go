package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/MikeKirillov/rocketservices/iam/internal/app"
	"github.com/MikeKirillov/rocketservices/iam/internal/config"
)

func main() {
	ctx := context.Background()

	// Загружаем переменные окружения из .env
	if err := godotenv.Load("../iam.env"); err != nil {
		slog.Error("ошибка загрузки переменных окружения из inventory.env", "error", err)
		os.Exit(1)
	}

	configPath := config.ResolveConfigPath()
	a := app.New(ctx, configPath)

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
		os.Exit(1)
	}
}

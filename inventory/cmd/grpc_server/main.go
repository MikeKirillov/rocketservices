package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/MikeKirillov/rocketservices/inventory/internal/app"
	"github.com/MikeKirillov/rocketservices/inventory/internal/config"
)

func main() {
	ctx := context.Background()

	// Загружаем переменные окружения из .env
	err := godotenv.Load("../inventory.env")
	if err != nil {
		slog.Error("ошибка загрузки переменных окружения из inventory.env", "error", err)
		os.Exit(1)
	}

	configPath := config.ResolveConfigPath()
	a := app.New(ctx, configPath)

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
		os.Exit(1)
	}

	// Подключаемся к PostgreSQL
	// dbURI := os.Getenv("DB_URI")
	// if dbURI == "" {
	// 	slog.Error("переменная окружения DB_URI не установлена")
	// 	os.Exit(1)
	// }
}

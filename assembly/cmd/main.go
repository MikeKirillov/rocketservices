package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/MikeKirillov/rocketservices/assembly/internal/app"
	"github.com/MikeKirillov/rocketservices/assembly/internal/config"
)

func main() {
	ctx := context.Background()

	configPath := config.ResolveConfigPath()
	a := app.New(ctx, configPath)

	if err := a.Run(); err != nil {
		slog.Error("ошибка при работе приложения", "error", err)
		os.Exit(1)
	}
}

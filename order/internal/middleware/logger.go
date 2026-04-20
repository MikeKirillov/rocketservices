package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Засекаем время начала обработки запроса
		start := time.Now()

		// Логируем начало запроса
		slog.Info("начало запроса", "method", r.Method, "path", r.URL.Path)

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)

		// Логируем окончание запроса с временем выполнения
		slog.Info("запрос завершен", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}

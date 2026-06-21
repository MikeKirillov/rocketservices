package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Описываем обработчик панических ситуаций
		defer func() {
			if re := recover(); re != nil {
				slog.Error("паника в HTTP методе",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Any("panic", re),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()

		// Передаем управление следующему обработчику
		next.ServeHTTP(w, r)
	})
}

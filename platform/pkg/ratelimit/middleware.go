package ratelimit

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

const (
	check_error    = "ошибка проверки rate limit"
	declined_error = "запрос отклонен rate limiter"
)

func Middleware(
	ctx context.Context,
	limiter *redis_rate.Limiter,
	rate, burst int,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// В качестве ключа — путь запроса (r.URL.Path).
		// Библиотека redis_rate сама добавляет префикс "rate:" к ключу (константа redisPrefix),
		// поэтому дополнительный префикс не нужен
		key := r.URL.Path
		res, err := limiter.Allow(ctx, key, redis_rate.Limit{Rate: rate, Burst: burst, Period: time.Second})
		if err != nil {
			slog.Error(check_error, "path", key, "error", err)
			return
		}

		// Если лимит превышен (res.Allowed == 0) — возвращайте http.StatusTooManyRequests (429)
		if res.Allowed == 0 {
			slog.Warn(
				declined_error,
				"path", key,
				"retry_after", res.RetryAfter,
			)
			http.Error(w, declined_error, http.StatusTooManyRequests)
		}

		// Fail-open стратегия: если Redis недоступен (ошибка от limiter.Allow) — пропускаем запрос дальше
		next.ServeHTTP(w, r)
	})
}

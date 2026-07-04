package ratelimit

import (
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"github.com/go-redis/redis_rate/v10"
)

const (
	normalizedError = "ошибка нормализации r.URL.Path"
	checkError      = "ошибка проверки rate limit"
	declinedError   = "запрос отклонен rate limiter"
	regexUuid       = `[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`
)

var uuidRegex = regexp.MustCompile(regexUuid)

func Middleware(
	limiter *redis_rate.Limiter,
	rate, burst int,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// В качестве ключа — путь запроса (r.URL.Path).
		// Библиотека redis_rate сама добавляет префикс "rate:" к ключу (константа redisPrefix),
		// поэтому дополнительный префикс не нужен
		path := r.URL.Path

		// нормализация запросов с указанием идентификаторов для приведения к типу "/api/v1/orders/{uuid}",
		// что дает избежать использования уникального заказа отдельным ключом в редисе
		key := uuidRegex.ReplaceAllString(path, "{uuid}")
		res, err := limiter.Allow(
			r.Context(),
			key,
			redis_rate.Limit{
				Rate:   rate,
				Burst:  burst,
				Period: time.Second,
			},
		)
		if err != nil {
			// Fail-open стратегия: если Redis недоступен (ошибка от limiter.Allow) — пропускаем запрос дальше
			slog.Error(checkError, "path", key, "error", err)
			next.ServeHTTP(w, r)
			return
		}

		// Если лимит превышен (res.Allowed == 0) — возвращайте http.StatusTooManyRequests (429)
		if res.Allowed == 0 {
			slog.Warn(
				declinedError,
				"path", key,
				"retry_after", res.RetryAfter,
			)
			w.Header().Set("Retry-After", time.Now().Add(res.RetryAfter).Format(time.RFC1123))
			http.Error(w, declinedError, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/MikeKirillov/rocketservices/order/internal/service/order"
	"github.com/MikeKirillov/rocketservices/platform/pkg/auth"
)

func Auth(next http.Handler, authClient order.AuthServiceClient) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "отсутствует заголовок Authorization", http.StatusUnauthorized)
		}

		if strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "формат: Bearer <session_uuid>", http.StatusUnauthorized)
		}

		sessionUUID := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionUUID == "" {
			http.Error(w, "пустой session_uuid", http.StatusUnauthorized)
		}

		sID, uID, err := authClient.Whoami(r.Context(), uuid.MustParse(sessionUUID))
		if err != nil {
			http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
		}

		ctx := auth.WithSessionUUID(r.Context(), sID.String())
		ctx = auth.WithUserUUID(ctx, uID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

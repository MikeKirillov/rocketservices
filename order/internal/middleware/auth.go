package middleware

import (
	"net/http"
	"strings"

	authClientV1 "github.com/MikeKirillov/rocketservices/order/internal/client/grpc/auth/v1"
	"github.com/MikeKirillov/rocketservices/platform/pkg/auth"
)

func Auth(next http.Handler, authClient *authClientV1.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "отсутствует заголовок Authorization", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "формат: Bearer <session_uuid>", http.StatusUnauthorized)
			return
		}

		sessionUUID := strings.TrimPrefix(authHeader, "Bearer ")
		if sessionUUID == "" {
			http.Error(w, "пустой session_uuid", http.StatusUnauthorized)
			return
		}

		sID, uID, err := authClient.Whoami(r.Context(), sessionUUID)
		if err != nil {
			http.Error(w, "ошибка авторизации", http.StatusUnauthorized)
			return
		}

		ctx := auth.WithSessionUUID(r.Context(), sID.String())
		ctx = auth.WithUserUUID(ctx, uID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

package middleware

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

type claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func (m Middleware) WithJwt(h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// Извлекаем из заголовка токен
		tokenStr := r.Header.Get("Authorization")

		// Если его нет, ошибка
		if tokenStr == "" {
			m.logger.Error("Header Authorization was not provided")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := &claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims,
			func(t *jwt.Token) (interface{}, error) {
				return []byte(m.cfg.SecretKey), nil
			})
		if err != nil {
			m.logger.Error("Something went wrong with token", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			m.logger.Error("Invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		m.logger.Debug("User id is ", claims.UserID, " set to ctx")

		// Устанавливаем идентификатор пользователя в контекст
		ctx := m.service.SetUserIdToCtx(r.Context(), int64(claims.UserID))
		r = r.WithContext(ctx)

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(fn)
}

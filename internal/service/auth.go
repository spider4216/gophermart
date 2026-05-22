package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spider4216/gophermart/internal/models"
)

type (
	userIdKey string
)

const (
	userKey userIdKey = "user_id"
)

type claims struct {
	jwt.RegisteredClaims
	UserID int
}

func (s Service) SetUserIdToCtx(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, userKey, userId)
}

func (s Service) GetUserIdFromCtx(ctx context.Context) int64 {
	userId, ok := ctx.Value(userKey).(int64)

	if !ok {
		return 0
	}

	return userId
}

func (s Service) BuildJWTString(userId int64, secret string, exp time.Duration) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(exp)),
		},
		UserID: int(userId),
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (s Service) CheckPass(user *models.User, pass string) bool {
	hash := sha256.Sum256([]byte(pass))
	hashString := hex.EncodeToString(hash[:])

	return hashString == user.Password
}

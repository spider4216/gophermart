package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/golang/mock/gomock"
	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/mocks"
	"github.com/spider4216/gophermart/internal/models"
	"github.com/spider4216/gophermart/internal/repository"
	"github.com/spider4216/gophermart/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type claims struct {
	jwt.RegisteredClaims
	UserID int64
}

func TestSignUp(t *testing.T) {
	// Логер
	sugar := zap.NewExample().Sugar()

	// Создаем конфигурацию
	cfg := &config.Config{
		CtxTimeout:   3 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
		MaxBodySize:  2048,
		ExpToken:     1 * time.Minute,
	}

	// создаём gomock контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var userId int64 = 1
	var balanceId int64 = 1
	login := "test1"
	pass := "qwerty"

	// создаём объект-заглушку
	m := mocks.NewMockStorage(ctrl)
	tx := mocks.NewMockTx(ctrl)
	// Ожидание на создание пользователя
	m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(userId, nil)
	// Ожидание на открытие транзакции
	m.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	// Ожидание на фиксацию транзакции
	tx.EXPECT().Commit().Return(nil)
	// Ожидание на создание баланса
	m.EXPECT().CreateUserBalance(gomock.Any(), userId).Return(balanceId, nil)

	// Создаем репозиторий
	repo := repository.New(m)
	// Создаем сервис
	service := service.New(repo, sugar, cfg, nil)
	// Создаем обработчик
	handler := New(cfg, sugar, service)

	// Тело запроса
	req := models.SignInReq{
		Login: login,
		Pass:  pass,
	}

	reqJson, err := json.Marshal(req)
	assert.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, "/api/user/register", bytes.NewBuffer(reqJson))
	w := httptest.NewRecorder()

	handler.SignUp(w, r)
	res := w.Result()

	// HTTP статус ок
	assert.Equal(t, http.StatusOK, res.StatusCode)
	// Заголовок с токеном ок
	tokenStr := res.Header.Get("Authorization")
	assert.NotEmpty(t, tokenStr)

	// Парсим токен
	claims := &claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(to *jwt.Token) (interface{}, error) {
			return []byte(cfg.SecretKey), nil
		})

	require.NoError(t, err)

	// Валидируем токен
	assert.True(t, token.Valid)
	// Проверяем ID пользователя
	assert.Equal(t, claims.UserID, userId)
}

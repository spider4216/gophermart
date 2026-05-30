package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestGetUserBalanceAndWithdrawn(t *testing.T) {
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

	var balance float32 = 358

	data := &models.Balance{
		ID:      1,
		UserID:  1,
		Balance: balance,
	}

	var withdrawn float32 = 101

	// создаём объект-заглушку
	m := mocks.NewMockStorage(ctrl)
	// Ожидание на создание пользователя
	m.EXPECT().GetUserBalance(gomock.Any(), gomock.Any()).Return(data, nil)
	m.EXPECT().GetTotalUserWithdrawn(gomock.Any(), gomock.Any()).Return(withdrawn, nil)

	// Создаем репозиторий
	repo := repository.New(m)
	// Создаем сервис
	service := service.New(repo, sugar, cfg, nil)
	// Создаем обработчик
	handler := New(cfg, sugar, service)

	r := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
	w := httptest.NewRecorder()

	handler.GetUserBalanceAndWithdrawn(w, r)
	res := w.Result()

	var body []byte

	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)

	var resp models.BalanceResp

	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)

	assert.Equal(t, resp.Current, balance)
	assert.Equal(t, resp.Withdrawn, withdrawn)
}

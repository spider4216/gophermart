package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestGetGetUserOrders(t *testing.T) {
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

	num := 3530111333300000

	data := []models.Order{
		{
			ID:        1,
			UserID:    1,
			Num:       num,
			Status:    models.OrderStatusProcessed,
			Accrual:   50,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	// создаём объект-заглушку
	m := mocks.NewMockStorage(ctrl)
	// Ожидание на создание пользователя
	m.EXPECT().GetUserOrders(gomock.Any(), gomock.Any()).Return(data, nil)

	// Создаем репозиторий
	repo := repository.New(m)
	// Создаем сервис
	service := service.New(repo, sugar, cfg, nil)
	// Создаем обработчик
	handler := New(cfg, sugar, service)

	r := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
	w := httptest.NewRecorder()

	handler.GetUserOrders(w, r)
	res := w.Result()

	var body []byte

	assert.Equal(t, res.StatusCode, http.StatusOK)

	body, err := io.ReadAll(res.Body)

	require.NoError(t, err)

	var resp []models.OrderResp

	err = json.Unmarshal(body, &resp)
	require.NoError(t, err)

	assert.Len(t, resp, 1)
	assert.Equal(t, resp[0].Number, strconv.Itoa(num))
}

package service

import (
	"context"
	"errors"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/repository"
	"github.com/theplant/luhn"
	"go.uber.org/zap"
)

func New(repo *repository.Repository, logger *zap.SugaredLogger, cfg *config.Config, httpC *resty.Client) *Service {
	return &Service{
		repo:      repo,
		logger:    logger,
		cfg:       cfg,
		httpC:     httpC,
		pauseChan: make(chan struct{}), // Изначально открытый канал
	}
}

type Service struct {
	repo      *repository.Repository
	logger    *zap.SugaredLogger
	cfg       *config.Config
	httpC     *resty.Client
	pauseMu   sync.RWMutex
	pauseChan chan struct{} // Канал, который закроем, когда пауза завершается
	isPaused  bool          // Флаг, пригодится в горутинах, чтобы понимать нужно ли ожидать
}

type NoBalance struct{}

func (e *NoBalance) Error() string {
	return "No balance for operation"
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) IsErrAsDuplicate(err error) bool {
	var pgErr *pgconn.PgError

	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.UniqueViolation
}

func (s *Service) IsOrderNumValid(num int) bool {
	return luhn.Valid(num)
}

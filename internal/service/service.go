package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/models"
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

func (s *Service) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) IncreaseUserBalance(ctx context.Context, userId int64, amount float32) error {
	b, err := s.GetUserBalance(ctx, userId)
	if err != nil {
		return err
	}

	sum := b.Balance + amount

	return s.UpdateUserBalance(ctx, userId, sum)
}

func (s *Service) UpdateUserBalance(ctx context.Context, userId int64, amount float32) error {
	return s.repo.UpdateUserBalance(ctx, userId, amount)
}

func (s *Service) GetUserBalance(ctx context.Context, userId int64) (*models.Balance, error) {
	return s.repo.GetUserBalance(ctx, userId)
}

func (s *Service) CreateUserBalance(ctx context.Context, userId int64) (int64, error) {
	s.logger.Debug("Create balance for user ", userId)
	return s.repo.CreateUserBalance(ctx, userId)
}

func (s *Service) UpdateOrderInvalid(ctx context.Context, orderNum int, userId int64) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusInvalid, 0)
}

func (s *Service) UpdateOrderProcessed(ctx context.Context, orderNum int, userId int64, sum float32) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusProcessed, sum)
}

func (s *Service) UpdateOrderProcess(ctx context.Context, orderNum int, userId int64) error {
	return s.repo.UpdateOrderStatus(ctx, orderNum, userId, models.OrderStatusProcess, 0)
}

func (s *Service) GetOrdersByUserId(ctx context.Context, userId int64) ([]models.Order, error) {
	return s.repo.GetOrdersByUserId(ctx, userId)
}

func (s *Service) GetOrderByUserId(ctx context.Context, num int, userId int64) (*models.Order, error) {
	return s.repo.GetOrderByUserId(ctx, num, userId)
}

func (s *Service) CreateOrder(ctx context.Context, userId int64, num int) (int64, error) {
	return s.repo.CreateOrder(ctx, userId, num)
}

func (s *Service) SignUpUser(ctx context.Context, username string, pass string) (int64, error) {
	hash := sha256.Sum256([]byte(pass))

	hashString := hex.EncodeToString(hash[:])

	id, err := s.repo.CreateUser(ctx, username, hashString)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Service) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	return s.repo.GetUserByUsername(ctx, login)
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

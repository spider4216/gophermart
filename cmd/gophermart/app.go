package main

import (
	"flag"
	"fmt"

	"github.com/spider4216/gophermart/internal/config"
	"github.com/spider4216/gophermart/internal/store"
	"github.com/spider4216/gophermart/migrations"
	"go.uber.org/zap"
)

type app struct {
	cfg    *config.Config
	logger *zap.SugaredLogger
	flags  *flags
	store  store.Storage
}

type flags struct {
	Host        string
	LogLvl      string
	Dsn         string
	AccrualAddr string
}

func newApp() app {
	return app{}
}

func (app *app) Run() error {
	// Инициализируем флаги
	if err := app.initFlags(); err != nil {
		return err
	}

	// Инициализируем конфигурацию
	if err := app.initConfig(); err != nil {
		return err
	}

	// Инициализируем логгер
	if err := app.initLogger(); err != nil {
		return err
	}

	// Инициализация хранилища
	if err := app.initStore(); err != nil {
		return err
	}

	// Запуск миграций
	if err := app.initMigrations(); err != nil {
		return err
	}

	return nil
}

func (app *app) initFlags() error {
	app.flags = &flags{}

	host := flag.String("a", "", "Net address host:port")
	logLvl := flag.String("l", "", "Log level: debug, info, warning, error, fatal")
	dsn := flag.String("d", "", "DSN")
	accAddr := flag.String("r", "", "Accrual Address")

	flag.Parse()

	app.flags.Host = *host
	app.flags.LogLvl = *logLvl
	app.flags.Dsn = *dsn
	app.flags.AccrualAddr = *accAddr

	return nil
}

func (app *app) initConfig() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	if cfg.RunAddress == "" {
		cfg.RunAddress = app.flags.Host
	}

	if cfg.Dsn == "" {
		cfg.Dsn = app.flags.Dsn
	}

	if cfg.LogLvl == "" {
		cfg.LogLvl = app.flags.LogLvl
	}

	if cfg.AccrualAddr == "" {
		cfg.AccrualAddr = app.flags.AccrualAddr
	}

	app.cfg = cfg

	return nil
}

func (app *app) initLogger() error {
	level, err := zap.ParseAtomicLevel(app.cfg.LogLvl)
	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = level

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	sugar := logger.Sugar()

	app.logger = sugar

	return nil
}

func (app *app) initStore() error {
	store, err := store.New(store.PostgreDriver, app.cfg.Dsn, app.logger)
	if err != nil {
		return err
	}

	app.store = store

	return nil
}

func (app *app) initMigrations() error {
	app.logger.Debug("Up migrations")

	st, ok := app.store.(*store.PgxStore)

	if !ok {
		return fmt.Errorf("cannot cast to pgx store type in init migration")
	}

	if err := migrations.Run(st.DB); err != nil {
		return err
	}

	return nil
}

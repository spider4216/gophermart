package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	RunAddress   string        `env:"RUN_ADDRESS" envDefault:":8081"`
	Dsn          string        `env:"DATABASE_URI"`
	AccrualAddr  string        `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://127.0.0.1:8080"`
	LogLvl       string        `env:"LOG_LVL" envDefault:"debug"`
	CtxTimeout   time.Duration `env:"CTX_TIMEOUT" envDefault:"3s"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
	IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"30s"`
	MaxBodySize  int64         `env:"MAX_BODY_SIZE" envDefault:"2048"`
}

func New() (*Config, error) {
	cfg := &Config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

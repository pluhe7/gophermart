package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pluhe7/gophermart/internal/logger"
)

const (
	defaultAddress  = ":8080"
	defaultLogLevel = "info"
)

type Config struct {
	Address              string `env:"RUN_ADDRESS" short:"a" long:"address" description:"server address; example: -a localhost:8080"`
	LogLevel             string `env:"LOG_LEVEL" short:"l" long:"log-level" description:"log level; example: -l debug"`
	DatabaseDSN          string `env:"DATABASE_URI" short:"d" long:"database" description:"data source name for db; example: -d host=host port=port user=myuser password=xxxx dbname=mydb sslmode=disable"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" short:"r" long:"accrual" description:"accrual server address; example: -a localhost:1234"`
}

func (cfg *Config) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("address", cfg.Address)
	encoder.AddString("log level", cfg.LogLevel)
	encoder.AddString("database dsn", cfg.DatabaseDSN)
	encoder.AddString("accrual system address", cfg.AccrualSystemAddress)

	return nil
}

func InitConfig() *Config {
	var cfg Config

	err := cfg.parseFlags()
	if err != nil {
		logger.Log.Fatal("parse flags", zap.Error(err))
	}

	err = cfg.parseEnv()
	if err != nil {
		logger.Log.Fatal("parse env variables", zap.Error(err))
	}

	cfg.fillEmptyWithDefault()

	return &cfg
}

func (cfg *Config) parseFlags() error {
	p := flags.NewParser(cfg, flags.IgnoreUnknown)
	_, err := p.Parse()

	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	return err
}

func (cfg *Config) parseEnv() error {
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("parse env: %w", err)
	}

	return err
}

func (cfg *Config) fillEmptyWithDefault() {
	if cfg.Address == "" {
		cfg.Address = defaultAddress
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = defaultLogLevel
	}
}

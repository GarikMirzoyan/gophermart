package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	ErrMissingConfig = errors.New("missing required configuration")
)

type Config struct {
	RunAddress     string
	DatabaseURI    string
	AccrualAddress string
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", getEnv("RUN_ADDRESS", ":8080"), "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnv("DATABASE_URI", ""), "database URI")
	flag.StringVar(&cfg.AccrualAddress, "r", getEnv("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	flag.Parse()

	if cfg.DatabaseURI == "" {
		return nil, fmt.Errorf("%w: DATABASE_URI is required", ErrMissingConfig)
	}
	if cfg.AccrualAddress == "" {
		return nil, fmt.Errorf("%w: ACCRUAL_SYSTEM_ADDRESS is required", ErrMissingConfig)
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

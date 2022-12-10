package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
)

type Config struct {
	APIConfig      APIConfig
	BatchConfig    BatchConfig
	DatabaseConfig DatabaseConfig
	TraceConfig    TraceConfig
}

type APIConfig struct {
	Addr string `env:"ADDR"`
}

type BatchConfig struct {
	SleepInterval time.Duration `env:"BATCH_SLEEP_INTERVAL"`
}

type TraceConfig struct {
	TraceURL         string `env:"TRACE_URL"`
	TraceServiceName string `env:"TRACE_SERVICE_NAME"`
}

type DatabaseConfig struct {
	Driver   string `env:"DRIVER" envDefault:"postgres"`
	Host     string `env:"PSQL_HOST,required"`
	Port     string `env:"PSQL_PORT,required"`
	User     string `env:"PSQL_USER,required"`
	Password string `env:"PSQL_PASSWORD,required"`
	Database string `env:"PSQL_NAME,required"`
}

func LoadConfig() (*Config, error) {
	var conf Config

	loadConfigFuncs := []func() error{
		func() error { return env.Parse(&conf.APIConfig) },
		func() error { return env.Parse(&conf.BatchConfig) },
		func() error { return env.Parse(&conf.DatabaseConfig) },
		func() error { return env.Parse(&conf.TraceConfig) },
	}

	for _, f := range loadConfigFuncs {
		if err := f(); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	// data, err := json.Marshal(conf)
	// fmt.Println(string(data), err)
	// return

	return &conf, nil
}

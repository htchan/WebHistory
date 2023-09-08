package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type APIConfig struct {
	BinConfig         APIBinConfig
	DatabaseConfig    DatabaseConfig
	TraceConfig       TraceConfig
	UserServiceConfig UserServiceConfig
	WebsiteConfig     WebsiteConfig
}

type APIBinConfig struct {
	Addr           string        `env:"ADDR"`
	ReadTimeout    time.Duration `env:"API_READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout   time.Duration `env:"API_WRITE_TIMEOUT" envDefault:"5s"`
	IdleTimeout    time.Duration `env:"API_IDLE_TIMEOUT" envDefault:"5s"`
	APIRoutePrefix string        `env:"WEB_WATCHER_API_ROUTE_PREFIX" envDefault:"/api/web-watcher"`
}

type BatchConfig struct {
	BinConfig      BatchBinConfig
	DatabaseConfig DatabaseConfig
	TraceConfig    TraceConfig
	WebsiteConfig  WebsiteConfig
}

type BatchBinConfig struct {
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

type UserServiceConfig struct {
	Addr  string `env:"USER_SERVICE_ADDR,required"`
	Token string `env:"USER_SERVICE_TOKEN,required"`
}

type WebsiteConfig struct {
	Separator     string `env:"WEB_WATCHER_SEPARATOR" envDefault:"\n"`
	MaxDateLength int    `env:"WEB_WATCHER_DATE_MAX_LENGTH" envDefault:"2"`
}

func LoadAPIConfig() (*APIConfig, error) {
	var conf APIConfig

	loadConfigFuncs := []func() error{
		func() error { return env.Parse(&conf) },
		func() error { return env.Parse(&conf.BinConfig) },
		func() error { return env.Parse(&conf.DatabaseConfig) },
		func() error { return env.Parse(&conf.TraceConfig) },
		func() error { return env.Parse(&conf.UserServiceConfig) },
		func() error { return env.Parse(&conf.WebsiteConfig) },
	}

	for _, f := range loadConfigFuncs {
		if err := f(); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	return &conf, nil
}

func LoadBatchConfig() (*BatchConfig, error) {
	var conf BatchConfig

	loadConfigFuncs := []func() error{
		func() error { return env.Parse(&conf) },
		func() error { return env.Parse(&conf.BinConfig) },
		func() error { return env.Parse(&conf.DatabaseConfig) },
		func() error { return env.Parse(&conf.TraceConfig) },
		func() error { return env.Parse(&conf.WebsiteConfig) },
	}

	for _, f := range loadConfigFuncs {
		if err := f(); err != nil {
			return nil, fmt.Errorf("parse config: %w", err)
		}
	}

	return &conf, nil
}

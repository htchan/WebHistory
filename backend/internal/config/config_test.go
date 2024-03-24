package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	leak := flag.Bool("leak", false, "check for memory leaks")
	flag.Parse()

	if *leak {
		goleak.VerifyTestMain(m)
	} else {
		os.Exit(m.Run())
	}
}

func Test_LoadAPIConfig(t *testing.T) {
	tests := []struct {
		name         string
		envMap       map[string]string
		expectedConf *APIConfig
		expectError  bool
	}{
		{
			name: "happy flow with default",
			envMap: map[string]string{
				"PSQL_HOST":          "host",
				"PSQL_PORT":          "port",
				"PSQL_USER":          "user",
				"PSQL_PASSWORD":      "password",
				"PSQL_NAME":          "name",
				"USER_SERVICE_ADDR":  "user_serv_addr",
				"USER_SERVICE_TOKEN": "user_serv_token",
			},
			expectedConf: &APIConfig{
				BinConfig: APIBinConfig{
					ReadTimeout:    5 * time.Second,
					WriteTimeout:   5 * time.Second,
					IdleTimeout:    5 * time.Second,
					APIRoutePrefix: "/api/web-watcher",
				},
				DatabaseConfig: DatabaseConfig{
					Driver:   "postgres",
					Host:     "host",
					Port:     "port",
					User:     "user",
					Password: "password",
					Database: "name",
				},
				UserServiceConfig: UserServiceConfig{
					Addr: "user_serv_addr", Token: "user_serv_token",
				},
				WebsiteConfig: WebsiteConfig{
					Separator:     "\n",
					MaxDateLength: 2,
				},
			},
			expectError: false,
		},
		{
			name: "happy flow without default",
			envMap: map[string]string{
				"WEB_WATCHER_SEPARATOR":        ",",
				"WEB_WATCHER_DATE_MAX_LENGTH":  "10",
				"ADDR":                         "addr",
				"API_READ_TIMEOUT":             "1s",
				"API_WRITE_TIMEOUT":            "1s",
				"API_IDLE_TIMEOUT":             "1s",
				"WEB_WATCHER_API_ROUTE_PREFIX": "prefix",
				"TRACE_URL":                    "trace_url",
				"TRACE_SERVICE_NAME":           "trace_service_name",
				"DRIVER":                       "driver",
				"PSQL_HOST":                    "host",
				"PSQL_PORT":                    "port",
				"PSQL_USER":                    "user",
				"PSQL_PASSWORD":                "password",
				"PSQL_NAME":                    "name",
				"USER_SERVICE_ADDR":            "user_serv_addr",
				"USER_SERVICE_TOKEN":           "user_serv_token",
			},
			expectedConf: &APIConfig{
				BinConfig: APIBinConfig{
					Addr:           "addr",
					ReadTimeout:    1 * time.Second,
					WriteTimeout:   1 * time.Second,
					IdleTimeout:    1 * time.Second,
					APIRoutePrefix: "prefix",
				},
				TraceConfig: TraceConfig{
					TraceURL:         "trace_url",
					TraceServiceName: "trace_service_name",
				},
				DatabaseConfig: DatabaseConfig{
					Driver:   "driver",
					Host:     "host",
					Port:     "port",
					User:     "user",
					Password: "password",
					Database: "name",
				},
				UserServiceConfig: UserServiceConfig{
					Addr: "user_serv_addr", Token: "user_serv_token",
				},
				WebsiteConfig: WebsiteConfig{
					Separator:     ",",
					MaxDateLength: 10,
				},
			},
			expectError: false,
		},
		{
			name:         "not providing required error",
			envMap:       map[string]string{},
			expectedConf: nil,
			expectError:  true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// populate env
			for key, value := range test.envMap {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			conf, err := LoadAPIConfig()
			assert.Equal(t, test.expectedConf, conf)
			assert.Equal(t, test.expectError, (err != nil))
		})
	}
}

func Test_LoadWorkerConfig(t *testing.T) {
	tests := []struct {
		name         string
		envMap       map[string]string
		expectedConf *WorkerConfig
		expectError  bool
	}{
		{
			name: "happy flow with default",
			envMap: map[string]string{
				"WEBSITE_UPDATE_SLEEP_INTERVAL": "10s",
				"WORKER_EXECUTOR_COUNT":         "10",
				"PSQL_HOST":                     "host",
				"PSQL_PORT":                     "port",
				"PSQL_USER":                     "user",
				"PSQL_PASSWORD":                 "password",
				"PSQL_NAME":                     "name",
				"REDIS_URL":                     "localhost:6543",
			},
			expectedConf: &WorkerConfig{
				BinConfig: WorkerBinConfig{
					WebsiteUpdateSleepInterval: 10 * time.Second,
					WorkerExecutorCount:        10,
					SupportHosts:               nil,
				},
				DatabaseConfig: DatabaseConfig{
					Driver:   "postgres",
					Host:     "host",
					Port:     "port",
					User:     "user",
					Password: "password",
					Database: "name",
				},
				WebsiteConfig: WebsiteConfig{
					Separator:     "\n",
					MaxDateLength: 2,
				},
				RedisURL: "localhost:6543",
			},
			expectError: false,
		},
		{
			name: "happy flow without default",
			envMap: map[string]string{
				"WEB_WATCHER_SEPARATOR":         ",",
				"WEB_WATCHER_DATE_MAX_LENGTH":   "10",
				"WEBSITE_UPDATE_SLEEP_INTERVAL": "10s",
				"WORKER_EXECUTOR_COUNT":         "10",
				"TRACE_URL":                     "trace_url",
				"TRACE_SERVICE_NAME":            "trace_service_name",
				"DRIVER":                        "driver",
				"PSQL_HOST":                     "host",
				"PSQL_PORT":                     "port",
				"PSQL_USER":                     "user",
				"PSQL_PASSWORD":                 "password",
				"PSQL_NAME":                     "name",
				"REDIS_URL":                     "localhost:6543",
				"SUPPORT_HOSTS":                 "host1,host2",
			},
			expectedConf: &WorkerConfig{
				BinConfig: WorkerBinConfig{
					WebsiteUpdateSleepInterval: 10 * time.Second,
					WorkerExecutorCount:        10,
					SupportHosts:               []string{"host1", "host2"},
				},
				TraceConfig: TraceConfig{
					TraceURL:         "trace_url",
					TraceServiceName: "trace_service_name",
				},
				DatabaseConfig: DatabaseConfig{
					Driver:   "driver",
					Host:     "host",
					Port:     "port",
					User:     "user",
					Password: "password",
					Database: "name",
				},
				WebsiteConfig: WebsiteConfig{
					Separator:     ",",
					MaxDateLength: 10,
				},
				RedisURL: "localhost:6543",
			},
			expectError: false,
		},
		{
			name:         "not providing required error",
			envMap:       map[string]string{},
			expectedConf: nil,
			expectError:  true,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			// populate env
			for key, value := range test.envMap {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			conf, err := LoadWorkerConfig()
			assert.Equal(t, test.expectedConf, conf)
			assert.Equal(t, test.expectError, (err != nil))
		})
	}
}

package config

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func Test_LoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		envMap       map[string]string
		expectedConf *Config
		expectError  bool
	}{
		{
			name: "happy flow with default",
			envMap: map[string]string{
				"BACKUP_DIRECTORY":   "backup_dir",
				"PSQL_HOST":          "host",
				"PSQL_PORT":          "port",
				"PSQL_USER":          "user",
				"PSQL_PASSWORD":      "password",
				"PSQL_NAME":          "name",
				"USER_SERVICE_ADDR":  "user_serv_addr",
				"USER_SERVICE_TOKEN": "user_serv_token",
			},
			expectedConf: &Config{
				APIConfig: APIConfig{
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
				BackupDirectory: "backup_dir",
				Separator:       "\n",
				MaxDateLength:   2,
			},
			expectError: false,
		},
		{
			name: "happy flow without default",
			envMap: map[string]string{
				"BACKUP_DIRECTORY":             "backup_dir",
				"WEB_WATCHER_SEPARATOR":        ",",
				"WEB_WATCHER_DATE_MAX_LENGTH":  "10",
				"ADDR":                         "addr",
				"API_READ_TIMEOUT":             "1s",
				"API_WRITE_TIMEOUT":            "1s",
				"API_IDLE_TIMEOUT":             "1s",
				"WEB_WATCHER_API_ROUTE_PREFIX": "prefix",
				"BATCH_SLEEP_INTERVAL":         "10s",
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
			expectedConf: &Config{
				APIConfig: APIConfig{
					Addr:           "addr",
					ReadTimeout:    1 * time.Second,
					WriteTimeout:   1 * time.Second,
					IdleTimeout:    1 * time.Second,
					APIRoutePrefix: "prefix",
				},
				BatchConfig: BatchConfig{SleepInterval: 10 * time.Second},
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
				BackupDirectory: "backup_dir",
				Separator:       ",",
				MaxDateLength:   10,
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

			conf, err := LoadConfig()
			if (err != nil) != test.expectError {
				t.Error("error diff")
				t.Errorf("expected: %v", test.expectError)
				t.Errorf("actual  : %v", err)
			}

			if !cmp.Equal(test.expectedConf, conf) {
				t.Errorf("conf diff: %v", cmp.Diff(test.expectedConf, conf))
			}
		})
	}
}

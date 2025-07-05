package env

import (
	"github.com/pvxdv/self_improver/internal/config/telegram"
	"os"
	"strconv"
	"time"

	"github.com/pvxdv/self_improver/internal/config"
	"github.com/pvxdv/self_improver/internal/config/app"
	"github.com/pvxdv/self_improver/internal/config/http"
	"github.com/pvxdv/self_improver/internal/config/storage"
)

const (
	AppEnvKey   = "APP_ENV"
	AppDebugKey = "APP_DEBUG"

	HTTPPortKey    = "HTTP_PORT"
	TimeoutKey     = "TIMEOUT"
	IdleTimeoutKey = "IDLE_TIMEOUT"

	DBHostKey     = "DB_HOST"
	DBPortKey     = "DB_PORT"
	DBUserKey     = "DB_USER"
	DBPasswordKey = "DB_PASSWORD"
	DBNameKey     = "DB_NAME"

	TToken    = "TELEGRAM_TOKEN"
	TPassword = "TELEGRAM_PASSWORD"
)

type LoaderEnv struct{}

func NewLoader() *LoaderEnv {
	return &LoaderEnv{}
}

func (l *LoaderEnv) Load() (*config.Config, error) {
	appCfg, err := loadAppConfig()
	if err != nil {
		return nil, err
	}

	httpCfg, err := loadHTTPConfig()
	if err != nil {
		return nil, err
	}

	dbCfg, err := loadDatabaseConfig()
	if err != nil {
		return nil, err
	}

	tCfg, err := loadTelegramConfig()
	if err != nil {
		return nil, err
	}

	return &config.Config{
		App:      appCfg,
		HTTP:     httpCfg,
		Storage:  dbCfg,
		Telegram: tCfg,
	}, nil
}

func loadAppConfig() (*app.Config, error) {
	env := os.Getenv(AppEnvKey)
	debugStr := os.Getenv(AppDebugKey)
	debug, _ := strconv.ParseBool(debugStr)

	return &app.Config{Env: env, Debug: debug}, nil
}

func loadHTTPConfig() (*http.Config, error) {
	port := os.Getenv(HTTPPortKey)
	timeoutStr := os.Getenv(TimeoutKey)
	idleTimeoutStr := os.Getenv(IdleTimeoutKey)

	var timeout time.Duration
	var idleTimeout time.Duration

	if timeoutStr != "" {
		t, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return nil, err
		}
		timeout = t
	}

	if idleTimeoutStr != "" {
		t, err := time.ParseDuration(idleTimeoutStr)
		if err != nil {
			return nil, err
		}
		idleTimeout = t
	}

	return &http.Config{
		Port:        port,
		Timeout:     timeout,
		IdleTimeout: idleTimeout,
	}, nil
}

func loadDatabaseConfig() (*storage.Config, error) {
	host := os.Getenv(DBHostKey)
	port := os.Getenv(DBPortKey)
	user := os.Getenv(DBUserKey)
	password := os.Getenv(DBPasswordKey)
	name := os.Getenv(DBNameKey)

	return &storage.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}, nil
}

func loadTelegramConfig() (*telegram.Config, error) {
	token := os.Getenv(TToken)
	pass := os.Getenv(TPassword)

	return &telegram.Config{
		Token: token, Password: pass,
	}, nil
}

package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App struct {
		Name string `envconfig:"APP_NAME"`
		Port int    `envconfig:"APP_PORT"`
		Env  string `envconfig:"APP_ENV"`
	}
	Logger struct {
		Mode  string `envconfig:"LOGGER_MODE"`
		Level string `envconfig:"LOGGER_LEVEL"`
	}
	Auth struct {
		AccessTokenPrivateKey string `envconfig:"AUTH_ACCESS_TOKEN_PRIVATE_KEY"`
		AccessTokenPublicKey  string `envconfig:"AUTH_ACCESS_TOKEN_PUBLIC_KEY"`
		AccessTokenSecretKey  string `envconfig:"AUTH_ACCESS_TOKEN_SECRET_KEY"`
		RefreshTokenSecretKey string `envconfig:"AUTH_REFRESH_TOKEN_SECRET_KEY"`
		AccessTokenExpireIn   int    `envconfig:"AUTH_ACCESS_TOKEN_EXPIRE_IN"`
		RefreshTokenExpireIn  int    `envconfig:"AUTH_REFRESH_TOKEN_EXPIRE_IN"`
	}
	DB struct {
		Host     string `envconfig:"DB_HOST"`
		Port     int    `envconfig:"DB_PORT"`
		User     string `envconfig:"DB_USER"`
		Password string `envconfig:"DB_PASSWORD"`
		DBName   string `envconfig:"DB_NAME"`
	}
	Graceful struct {
		Verbose               bool `envconfig:"GRACEFUL_VERBOSE"`
		StepDelay             int  `envconfig:"GRACEFUL_STEP_DELAY"`
		ServerShutdownTimeout int  `envconfig:"GRACEFUL_SERVER_SHUTDOWN_TIMEOUT"`
	}
	Vite struct {
		BaseURL string `envconfig:"VITE_API_BASE_URL"`
	}
}

func LoadConfig(envFiles ...string) (*Config, error) {
	err := godotenv.Load(envFiles...)
	if err != nil {
		return nil, err
	}

	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

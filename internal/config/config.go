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
		AppCode                 string `envconfig:"AUTH_APP_CODE" default:"isme"`
		EndpointWebSSOLogin     string `envconfig:"AUTH_ENDPOINT_WEB_SSO_LOGIN"`
		EndpointWebAcceptInvite string `envconfig:"AUTH_ENDPOINT_WEB_ACCEPT_INVITE"`
		AccessTokenPrivateKey   string `envconfig:"AUTH_ACCESS_TOKEN_PRIVATE_KEY"`
		AccessTokenPublicKey    string `envconfig:"AUTH_ACCESS_TOKEN_PUBLIC_KEY"`
		RefreshTokenSecretKey   string `envconfig:"AUTH_REFRESH_TOKEN_SECRET_KEY"`
		AccessTokenExpireIn     int    `envconfig:"AUTH_ACCESS_TOKEN_EXPIRE_IN"`
		RefreshTokenExpireIn    int    `envconfig:"AUTH_REFRESH_TOKEN_EXPIRE_IN"`
		ExternalLoginSessionTTL int    `envconfig:"AUTH_EXTERNAL_LOGIN_SESSION_TTL"`
		ExternalExchangeCodeTTL int    `envconfig:"AUTH_EXTERNAL_EXCHANGE_CODE_TTL"`
	}
	DB struct {
		// Driver selects the backend: "sqlite" (default) or "postgres". SQLite
		// stays the default so existing .env/prod is unaffected.
		Driver string `envconfig:"DB_DRIVER" default:"sqlite"`
		// SQLitePath is the SQLite file path (sqlite driver only).
		SQLitePath string `envconfig:"DB_SQLITE_PATH" default:"db/app.db"`
		// Host/Port/User/Password/DBName are the Postgres connection fields,
		// used when Driver=postgres and DSN is empty.
		Host     string `envconfig:"DB_HOST"`
		Port     int    `envconfig:"DB_PORT"`
		User     string `envconfig:"DB_USER"`
		Password string `envconfig:"DB_PASSWORD"`
		DBName   string `envconfig:"DB_NAME"`
		// SSLMode is the Postgres sslmode (default "disable").
		SSLMode string `envconfig:"DB_SSLMODE" default:"disable"`
		// DSN is an optional full Postgres DSN override; when set it takes
		// precedence over the discrete Host/Port/... fields.
		DSN string `envconfig:"DB_DSN" default:""`
	}
	Graceful struct {
		Verbose               bool `envconfig:"GRACEFUL_VERBOSE"`
		StepDelay             int  `envconfig:"GRACEFUL_STEP_DELAY"`
		ServerShutdownTimeout int  `envconfig:"GRACEFUL_SERVER_SHUTDOWN_TIMEOUT"`
	}
	Vite struct {
		BaseURL string `envconfig:"VITE_API_BASE_URL"`
	}
	AES struct {
		Secret string `envconfig:"AES_SECRET"`
	}
	Scheduler struct {
		// Master kill-switch for background schedulers (default true). When
		// false, the session auto-revoke job is never installed regardless of
		// the persisted DB config.
		Enabled bool `envconfig:"SCHEDULER_ENABLED" default:"true"`
	}
	Medioa struct {
		// BaseURL is the medioa2 origin used for server-to-server upload calls.
		// Use 127.0.0.1:<port> (not *.local) to avoid the ~5s mDNS resolver stall.
		BaseURL string `envconfig:"MEDIOA_BASE_URL" default:"http://127.0.0.1:8082"`
		// APIKey is the medioa2 public-API key (an "mk_..." value) sent as
		// X-API-Key on uploads. Server-side only — never exposed to the browser.
		// Minted by a member of the target medioa2 bucket; set in .env at runtime.
		APIKey string `envconfig:"MEDIOA_API_KEY"`
	}
}

func LoadConfig(envFiles ...string) (*Config, error) {
	// .env is optional — absent in deploy (fly.io etc.) where config is supplied
	// via real environment variables. A missing file is not fatal; envconfig reads
	// the OS environment below regardless.
	_ = godotenv.Load(envFiles...)

	cfg := new(Config)
	if err := envconfig.Process("", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

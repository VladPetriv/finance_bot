package config

import (
	"log"
	"sync"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents an app config.
type Config struct {
	App            App
	Telegram       Telegram
	PostgreSQL     PostgreSQL
	CurrencyBeacon CurrencyBeacon
	Gemini         Gemini
	Logger         Logger
}

// App represents an app configuration.
type App struct {
	OperationCreationInterval               time.Duration `env:"FB_APP_OPERATION_CREATION_INTERVAL" env-default:"5m"`
	ExtendingScheduledOperationsInterval    time.Duration `env:"FB_APP_EXTENDING_SCHEDULED_OPERATIONS_INTERVAL" env-default:"1h"`
	NotifyAboutSubscriptionPaymentsInterval time.Duration `env:"FB_APP_NOTIFY_ABOUT_SUBSCRIPTION_PAYMENTS_INTERVAL" env-default:"1m"`
}

// Telegram represents a telegram bot configuration.
type Telegram struct {
	BotToken     string `env:"FB_TELEGRAM_BOT_TOKEN"`
	WebhookURL   string `env:"FB_TELEGRAM_WEBHOOK_URL"`
	SeverAddress string `env:"FB_TELEGRAM_SERVER_ADDRESS" env-default:":8443"`
	UpdatesType  string `env:"FB_TELEGRAM_UPDATES_TYPE" env-default:"polling"`
}

// PostgreSQL represents a PostgreSQL database configuration.
type PostgreSQL struct {
	User     string `env:"FB_POSTGRESQL_USER" env-default:"root"`
	Password string `env:"FB_POSTGRESQL_PASSWORD" env-default:"admin"`
	Database string `env:"FB_POSTGRESQL_DATABASE" env-default:"finance_bot"`
	Host     string `env:"FB_POSTGRESQL_HOST" env-default:"localhost"`
	Port     string `env:"FB_POSTGRESQL_PORT" env-default:"5432"`
	SSLMode  string `env:"FB_POSTGRESQL_SSL_MODE" env-default:"disable"`
	URL      string `env:"FB_POSTGRESQL_URL"`
}

// CurrencyBeacon represents a config for CurrencyBeacon API.
type CurrencyBeacon struct {
	APIKey      string `env:"FB_CURRENCY_BEACON_API_KEY"`
	APIEndpoint string `env:"FB_CURRENCY_BEACON_API_ENDPOINT" env-default:"https://api.currencybeacon.com"`
}

// Gemini represents a config for Gemini API.
type Gemini struct {
	APIKey string `env:"FB_GEMINI_API_KEY"`
	Model  string `env:"FB_GEMINI_MODEL" env-default:"gemini-2.5-flash"`
}

// Logger represents a logger configuration.
type Logger struct {
	LogLevel        string `env:"FB_LOGGER_LOG_LEVEL" env-default:"debug"`
	LogFilename     string `env:"FB_LOGGER_LOG_FILENAME" env-default:""`
	PrettyLogOutput bool   `env:"FB_LOGGER_PRETTY_LOG_OUTPUT" env-default:"false"`
}

var (
	config Config
	once   sync.Once
)

// Get returns a new config.
func Get() *Config {
	once.Do(func() {
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			log.Fatalf("read env: %v", err)
		}
	})

	return &config
}

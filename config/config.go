package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config represents an app config.
type Config struct {
	Telegram Telegram
	MongoDB  MongoDB
	Logger   Logger
}

// Telegram represents a telegram bot configuration.
type Telegram struct {
	BotToken     string `env:"BOT_TOKEN"`
	WebhookURL   string `env:"WEBHOOK_URL"`
	SeverAddress string `env:"SERVER_ADDRESS" env-default:":8443"`
}

// MongoDB represents a mongoDB database configuration.
type MongoDB struct {
	URI      string `env:"MONGODB_URI" env-default:"mongodb://localhost:27017"`
	Database string `env:"MONGODB_DATABASE" env-default:"api"`
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

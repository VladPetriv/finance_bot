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
}

// Telegram represents Telegram bot configuration.
type Telegram struct {
	BotToken string `env:"BOT_TOKEN"`
}

// MongoDB represents MongoDB database configuration.
type MongoDB struct {
	URI      string `env:"MONGODB_URI" env-default:"mongodb://localhost:27017"`
	Database string `env:"MONGODB_DATABASE" env-default:"api"`
}

var (
	config Config
	once   sync.Once
)

// Get returns a Config.
func Get() *Config {
	once.Do(func() {
		err := cleanenv.ReadEnv(&config)
		if err != nil {
			log.Fatalf("read env: %v", err)
		}
	})

	return &config
}

package config

import (
	"log"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	cfg  Config
	once sync.Once
)

type Server struct {
	Host string `env:"HOST" envDefault:"localhost"`
	Port string `env:"PORT" envDefault:"3000"`
}

type Config struct {
	Server
}

func Load() *Config {
	once.Do(func() {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil && !os.IsNotExist(err) {
			log.Printf("failed to read .env: %v", err)
			panic(err)
		}

		if err := cleanenv.ReadEnv(&cfg); err != nil {
			log.Printf("failed to read env: %v", err)
			panic(err)
		}
	})

	return &cfg
}

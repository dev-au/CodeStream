package src

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strings"
)

var Config envData

type envData struct {
	Port            string   `env:"PORT" envDefault:"8000"`
	RedisUrl        string   `env:"REDIS_URL"`
	ApplicationMode string   `env:"APPLICATION_MODE"`
	Languages       []string `env:"LANGUAGES"`
}

func (envData) SetupEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Config = envData{
		Port:            os.Getenv("PORT"),
		RedisUrl:        os.Getenv("REDIS_URL"),
		ApplicationMode: os.Getenv("APPLICATION_MODE"),
		Languages:       strings.Split(os.Getenv("LANGUAGES"), ","),
	}
}

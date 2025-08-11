package src

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var Config envData

type envData struct {
	RedisUrl         string   `env:"REDIS_URL"`
	ApplicationMode  string   `env:"APPLICATION_MODE"`
	Languages        []string `env:"LANGUAGES"`
	CodeWorkDir      string   `env:"CODE_WORK_DIR"`
	RunTimeoutSecond int      `env:"RUN_TIMEOUT_SECOND"`
	GoogleCaptchaKey string   `env:"GOOGLE_CAPTCHA_KEY"`
}

func (envData) SetupEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	runTimeoutSecond, _ := strconv.Atoi(os.Getenv("RUN_TIMEOUT_SECOND"))

	Config = envData{
		RedisUrl:         os.Getenv("REDIS_URL"),
		ApplicationMode:  os.Getenv("APPLICATION_MODE"),
		Languages:        strings.Split(os.Getenv("LANGUAGES"), ","),
		CodeWorkDir:      os.Getenv("CODE_WORK_DIR"),
		RunTimeoutSecond: runTimeoutSecond,
		GoogleCaptchaKey: os.Getenv("GOOGLE_CAPTCHA_KEY"),
	}
}

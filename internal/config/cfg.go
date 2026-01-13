package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Tg TgBotCfg
}

type TgBotCfg struct {
	Token string
}

func LoadConfig() *Config {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatal(err, "cant't load config file")
		return nil
	}

	return cfg
}

func NewConfig() (*Config, error) {
	path := "/home/user/programmin/obsidianProject/.env"

	if err := godotenv.Load(path); err != nil {
		return nil, err

	}
	
	return &Config{
		Tg: TgBotCfg{Token: os.Getenv("TOKEN")},
	}, nil
}

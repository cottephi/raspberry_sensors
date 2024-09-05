package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/subosito/gotenv"
)

type Config struct {
	Api struct {
		Host string `yaml:"host" env:"SERVER_HOST" env-default:"localhost"`
		Port string `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
	} `yaml:"api"`
	Database struct {
		Host string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
		Port string `yaml:"port" env:"DB_PORT" env-default:"8080"`
		Token string `yaml:"token" env:"DB_TOKEN" env-required:"true"`
	} `yaml:"database"`
	Logger struct {
		Path  string `yaml:"path" env:"LOG_PATH" env-default:"./log.log"`
		Level string    `yaml:"level" env:"LOG_LEVEL" env-default:"INFO"`
	} `yaml:"logger"`
}

var once sync.Once
var config Config

func Get() Config {
	once.Do(func() {
		gotenv.Load()
		err := cleanenv.ReadConfig("config.yml", &config)
		if err != nil {
			panic(err)
		}
	})
	return config
}



package config

import (
	"fmt"
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
		Host  string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
		Port  string `yaml:"port" env:"DB_PORT" env-default:"8080"`
		Token string `yaml:"token" env:"DB_TOKEN" env-required:"true"`
	} `yaml:"database"`
	Logger struct {
		Path  string `env:"LOG_PATH" env-default:""`
		Level string `yaml:"level" env:"LOG_LEVEL" env-default:"INFO"`
	} `yaml:"logger"`
	Description string
}

func (c *Config) Validate() error {

	description := "Configuration:\n"

	nonEmptyStrings := [][2]string{
		{c.Api.Host, "Server Host"},
		{c.Api.Port, "Server Port"},
		{c.Database.Host, "Database Host"},
		{c.Database.Port, "Database Port"},
		{c.Logger.Level, "Logger Level"},
	}

	for _, value := range nonEmptyStrings {
		if value[0] == "" {
			return fmt.Errorf("Config %s can not be empty", value[1])
		}
		description += fmt.Sprintf(" - %s: %s\n", value[1], value[0])
	}
	description += " - Log file path: " + c.Logger.Path
	c.Description = description
	return nil
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
		err = config.Validate()
		if err != nil {
			panic(err)
		}
	})
	return config
}
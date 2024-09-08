package config

import (
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/subosito/gotenv"
)

type Config struct {
	Api struct {
		Host string `yaml:"host" env:"SERVER_HOST" env-default:"http://localhost"`
		Port string `yaml:"port" env:"SERVER_PORT" env-default:"8080"`
		URL *url.URL
	} `yaml:"api"`
	Database struct {
		Host  string `yaml:"host" env:"DB_HOST" env-default:"http://localhost"`
		Port  string `yaml:"port" env:"DB_PORT" env-default:"8080"`
		Token string `yaml:"token" env:"DB_TOKEN"`
		URL *url.URL
	} `yaml:"database"`
	Logger struct {
		Path  string `env:"LOG_PATH" env-default:""`
		Level string `yaml:"level" env:"LOG_LEVEL" env-default:"INFO"`
	} `yaml:"logger"`
	Description string
}

func (c *Config) Validate() error {

	description := "Configuration:\n"

	var nonEmptyStrings [][2]string

	if c.Database.Token == "" {
		nonEmptyStrings = [][2]string{
			{c.Api.Host, "Server Host"},
			{c.Api.Port, "Server Port"},
			{c.Logger.Level, "Logger Level"},
		}
	} else {
		nonEmptyStrings = [][2]string{
			{c.Api.Host, "Server Host"},
			{c.Api.Port, "Server Port"},
			{c.Database.Host, "Database Host"},
			{c.Database.Port, "Database Port"},
			{c.Logger.Level, "Logger Level"},
		}
	}

	for _, value := range nonEmptyStrings {
		if value[0] == "" {
			return fmt.Errorf("Config %s can not be empty", value[1])
		}
		description += fmt.Sprintf(" - %s: %s\n", value[1], value[0])
	}
	description += " - Log file path: " + c.Logger.Path + "\n"

	var err error

	if !strings.HasPrefix(c.Api.Host, "http://") && !strings.HasPrefix(c.Api.Host, "https://"){
		return fmt.Errorf("value for Server Host has no protocol scheme. Add 'http://' or 'https://'")
	}

	c.Api.URL, err = url.Parse(fmt.Sprintf("%s:%s", c.Api.Host, c.Api.Port))
	if err != nil {
		return fmt.Errorf("error parsing Server Host and Port into an URL: %s", err)
	}
	description += fmt.Sprintf(" - Server URL: %s\n", c.Api.URL)

	if c.Database.Token == "" {
		description += "No database token given, not writing data to database"
		c.Description = description
		return nil
	}

	if !strings.HasPrefix(c.Database.Host, "http://") && !strings.HasPrefix(c.Database.Host, "https://"){
		return fmt.Errorf("value for Database Host has no protocol scheme. Add 'http://' or 'https://'")
	}
	c.Database.URL, err = url.Parse(fmt.Sprintf("%s:%s", c.Api.Host, c.Api.Port))
	if err != nil {
		return fmt.Errorf("error parsing Database Host and Port into an URL: %s", err)
	}
	description += fmt.Sprintf(" - Database URL: %s\n", c.Api.URL)

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
package config

import (
	"fmt"
	"log"
	"sync"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
	Server        Server
	Database      Database
	CORS          CORS
	ConfluenceUrl string `env:"CONFLUENCE_URL"`
}

type Server struct {
	Hostname string `env:"HOSTNAME"`
	Port     string `env:"PORT,notEmpty"`
}

type Database struct {
	Host     string `env:"DB_HOST,notEmpty"`
	Port     string `env:"DB_PORT" envDefault:"5432"`
	User     string `env:"DB_USER,notEmpty"`
	Password string `env:"DB_PASSWORD,notEmpty"`
	Name     string `env:"DB_NAME,notEmpty"`
	SSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`
	TimeZone string `env:"DB_TIMEZONE" envDefault:"Asia/Bangkok"`
}

func (d Database) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode, d.TimeZone,
	)
}

type CORS struct {
	AllowOrigin string `env:"CORS_ALLOW_ORIGIN" envDefault:"*"`
}

var once sync.Once
var config Config

func prefix(e string) string {
	if e == "" {
		return ""
	}
	return fmt.Sprintf("%s_", e)
}

func C(envPrefix string) Config {
	once.Do(func() {
		opts := env.Options{
			// Prefix: prefix(envPrefix),
		}

		var err error
		config, err = parseEnv[Config](opts)
		if err != nil {
			log.Fatal(err)
		}
	})
	return config
}

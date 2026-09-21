package config

import (
	"fmt"
	"log/slog"
)

// Postgres holds the database connection settings
type Postgres struct {
	Host     string `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     string `envconfig:"POSTGRES_PORT" default:"5432"`
	User     string `envconfig:"POSTGRES_USER" default:"hotel"`
	Password string `envconfig:"POSTGRES_PASSWORD" default:"hotel"`
	Name     string `envconfig:"POSTGRES_DB_NAME" default:"hotel"`
}

func (p Postgres) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		p.Host, p.Port, p.User, p.Password, p.Name)
}

// Config holds every setting the service reads from its environment
type Config struct {
	Port     string     `envconfig:"PORT" default:"8080"`
	LogLevel slog.Level `envconfig:"LOG_LEVEL" default:"info"`
	Postgres Postgres
}

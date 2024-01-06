// config/config.go

package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config represents the configuration values.
type Config struct {
	Accept        string
	Authorization string
}

// LoadConfig loads the configuration from the environment file.
func LoadConfig() (*Config, error) {
	err := godotenv.Load(".envrc")
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file")
	}

	return &Config{
		Accept:        os.Getenv("ACCEPT"),
		Authorization: os.Getenv("AUTHORIZATION"),
	}, nil
}

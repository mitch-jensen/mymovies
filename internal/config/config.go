// Package config loads runtime configuration for the movie service.
package config

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"
)

// DbConfig contains database connection settings.
type DbConfig struct {
	Address  string
	Port     string
	User     string
	Password string
	Name     string
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Address string
	Port    string
}

// Load reads database and server configuration from the environment and .env.
func Load(configPath string) (*DbConfig, *ServerConfig, error) {
	config := viper.New()

	config.AutomaticEnv()

	_, err := os.Stat(".env")
	if err == nil {
		config.SetConfigType("env")
		config.AddConfigPath(configPath)
		config.SetConfigFile(".env")

		err = config.ReadInConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("read config: %w", err)
		}
	}

	return &DbConfig{
			Address:  config.GetString("POSTGRES_ADDRESS"),
			Port:     config.GetString("POSTGRES_PORT"),
			User:     config.GetString("POSTGRES_USER"),
			Password: config.GetString("POSTGRES_PASSWORD"),
			Name:     config.GetString("POSTGRES_DB"),
		}, &ServerConfig{
			Address: config.GetString("SERVER_ADDRESS"),
			Port:    config.GetString("SERVER_PORT"),
		}, nil
}

// ConnectionString returns a PostgreSQL connection URL for the database.
func (c *DbConfig) ConnectionString() string {
	hostPort := net.JoinHostPort(c.Address, c.Port)

	return fmt.Sprintf("postgresql://%s:%s@%s/%s",
		c.User, c.Password, hostPort, c.Name)
}

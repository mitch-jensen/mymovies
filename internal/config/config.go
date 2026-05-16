package config

import (
	"fmt"
	"net"
	"os"

	"github.com/spf13/viper"
)

type DbConfig struct {
	Address  string
	Port     string
	User     string
	Password string
	Name     string
}

type ServerConfig struct {
	Address string
	Port    string
}

func Load(configPath string) (*DbConfig, *ServerConfig, error) {
	v := viper.New()

	v.AutomaticEnv()

	if _, err := os.Stat(".env"); err == nil {
		v.SetConfigType("env")
		v.AddConfigPath(configPath)
		v.SetConfigFile(".env")
		if err := v.ReadInConfig(); err != nil {
			return nil, nil, fmt.Errorf("read config: %w", err)
		}
	}

	return &DbConfig{
			Address:  v.GetString("POSTGRES_ADDRESS"),
			Port:     v.GetString("POSTGRES_PORT"),
			User:     v.GetString("POSTGRES_USER"),
			Password: v.GetString("POSTGRES_PASSWORD"),
			Name:     v.GetString("POSTGRES_DB"),
		}, &ServerConfig{
			Address: v.GetString("SERVER_ADDRESS"),
			Port:    v.GetString("SERVER_PORT"),
		}, nil
}

func (c *DbConfig) ConnectionString() string {
	hostPort := net.JoinHostPort(c.Address, c.Port)
	return fmt.Sprintf("postgresql://%s:%s@%s/%s",
		c.User, c.Password, hostPort, c.Name)
}

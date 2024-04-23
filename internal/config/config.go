package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port     int    `yaml:"port"`
	Hostname string `yaml:"hostname"`
	ReadTimeout int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"database_name"`
}

func LoadConfig(filename string) (Config, error) {
	var config Config

	data, err := os.ReadFile(filename)
	if err != nil {
		return config, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return config, nil
}

func (c Config) CreateConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", 
		c.Database.Username, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.DatabaseName)
}

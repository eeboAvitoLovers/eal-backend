// Package config предоставляет функционал для загрузки и обработки конфигурационных файлов.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config содержит параметры конфигурации сервера и базы данных.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

// ServerConfig содержит параметры конфигурации сервера.
type ServerConfig struct {
	Port     int    `yaml:"port"`
	Hostname string `yaml:"hostname"`
	ReadTimeout int `yaml:"read_timeout"`
	WriteTimeout int `yaml:"write_timeout"`
}

// DatabaseConfig содержит параметры конфигурации базы данных.
type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"database_name"`
}

// LoadConfig загружает конфигурационный файл из указанного файла.
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

// CreateConnString создает строку подключения к базе данных на основе параметров конфигурации.
func (c Config) CreateConnString() string {
	if c.Database.Port != 0 {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", 
			c.Database.Username, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.DatabaseName)
	} else {
		return fmt.Sprintf("postgres://%s:%s@%s/%s", 
			c.Database.Username, c.Database.Password, c.Database.Host,  c.Database.DatabaseName)
	}
}
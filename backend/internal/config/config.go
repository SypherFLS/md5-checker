package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port    string `yaml:"port"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Sslmode  string `yaml:"sslmode"`
}

func InitConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return Config{}, err
	}

	var cfg Config

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.Database.Password = os.Getenv("DB_PASSWORD")

	return cfg, nil
}

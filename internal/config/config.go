package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Manager struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"manager"`

	Agent struct {
		ManagerURL string `yaml:"manager_url"`
	} `yaml:"agent"`

	Store struct {
		DBPath string `yaml:"db_path"`
	} `yaml:"store"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("internal/config/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func GetConfig() (*Config, error) {
	return Load()
}

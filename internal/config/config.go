package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	HttpHost string `yaml:"http_host"`
	HttpPort int    `yaml:"http_port"`
	GrpcHost string `yaml:"grpc_host"`
	GrpcPort int    `yaml:"grpc_port"`
}

type Agent struct {
	ManagerURL string `yaml:"manager_url"`
}

type Database struct {
	DBPath string `yaml:"db_path"`
}

type Security struct {
	CACertPath     string `yaml:"ca_cert_path"`
	CAKeyPath      string `yaml:"ca_key_path"`
	ServerCertPath string `yaml:"server_cert_path"`
	ServerKeyPath  string `yaml:"server_key_path"`
}

type Config struct {
	Manager Manager `yaml:"manager"`

	Agent Agent `yaml:"agent"`
	Database Database `yaml:"database"`

	ConfigPath string `yaml:"config_path"`
	StatePath  string `yaml:"state_path"`

	Security Security `yaml:"security"`
}

const (
	DefaultConfigPath = "/etc/mcloud/config.yaml"
)

func Load() (*Config, error) {
	data, err := os.ReadFile(DefaultConfigPath)
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

func SaveConfig(cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(DefaultConfigPath, data, 0644); err != nil {
		return err
	}

	return nil
}
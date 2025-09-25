package config

import (
	"gopkg.in/yaml.v3"

	"os"
	"path/filepath"
)

type Config struct {
	AppID         string         `yaml:"app-id"`
	Version       string         `yaml:"version"`
	Env           string         `yaml:"env"`
	Port          int            `yaml:"port"`
	StorageConfig *StorageConfig `yaml:"storage"`
	PubConfig     *PubConfig     `yaml:"pub"`
	SubConfig     *SubConfig     `yaml:"sub"`
}

type StorageConfig struct {
	Dsn string `yaml:"dsn"`
}

type PubConfig struct {
	RabbitURL  string `yaml:"rabbit-url"`
	Exchange   string `yaml:"exchange"`
	RoutingKey string `yaml:"routing-key"`
}

type SubConfig struct {
	Brokers     []string `yaml:"brokers"`
	GroupID     string   `yaml:"group-id"`
	Topics      []string `yaml:"topics"`
	StartOldest bool     `yaml:"start-oldest"`
}

func Parse(path string, file string) (*Config, error) {
	yamlFile, err := os.ReadFile(filepath.Join(path, file))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

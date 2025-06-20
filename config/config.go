package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	MongoDB MongoDBConfig `yaml:"mongo_db"`
}

type MongoDBConfig struct {
	URI string `yaml:"uri" required:"true"`
}

type ServerConfig struct {
	Host string `yaml:"host" required:"true"`
	Port string `yaml:"port" required:"true"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Package config provides functionality to load application configuration
// from a YAML file into a structured Go object.
package config

import (
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"
)

// Config holds all configurable fields for the application, including
// server and MongoDB connection settings.
type Config struct {
	Server  ServerConfig  `yaml:"server"`   // Server configuration (host, port)
	MongoDB MongoDBConfig `yaml:"mongo_db"` // MongoDB configuration (URI)
}

// MongoDBConfig contains the URI used to connect to the MongoDB instance.
type MongoDBConfig struct {
	URI string `yaml:"uri" required:"true"` // MongoDB connection URI
}

// ServerConfig defines the server's binding address and port.
type ServerConfig struct {
	Host string `yaml:"host" required:"true"` // Hostname or IP to bind the service
	Port string `yaml:"port" required:"true"` // Port on which the service listens
}

// LoadConfig reads and parses a YAML configuration file from the given path.
// It ensures the path is sanitized using filepath.Clean for security.
//
// Returns a pointer to the Config struct or an error if reading or parsing fails.
func LoadConfig(path string) (*Config, error) {
	safePath := filepath.Clean(path)
	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

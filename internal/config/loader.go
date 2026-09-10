package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gca-research-group/fabric-network-orchestrator/internal/validate"
	"gopkg.in/yaml.v3"
)

func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading config file: %w", err)
	}

	return loadConfig(data, strings.ToLower(filepath.Ext(path)))
}

// LoadConfigFromYAML parses and validates a YAML configuration held in memory.
func LoadConfigFromYAML(data []byte) (*Config, error) {
	return loadConfig(data, ".yaml")
}

func loadConfig(data []byte, extension string) (*Config, error) {
	var config Config
	var err error

	switch extension {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yml", ".yaml":
		err = yaml.Unmarshal(data, &config)
	case ".toml":
		err = toml.Unmarshal(data, &config)
	default:
		return nil, errors.New("unsupported config format")
	}

	if err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := validate.Config(config); err != nil {
		return nil, err
	}

	setUpDefaultValues(&config)

	return &config, nil
}

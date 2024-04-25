package config

import (
	"gopkg.in/yaml.v2"
	"os"
)

func LoadConfig(path string) (*Config, error) {
	configFile, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

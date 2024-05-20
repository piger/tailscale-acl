package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey  string `yaml:"api_key"`
	Tailnet string `yaml:"tailnet"`
}

func Read(filename string) (*Config, error) {
	fh, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	var config Config
	if err := yaml.NewDecoder(fh).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/kelseyhightower/envconfig"
)

var defaultYamlConfigPath = "./config.yaml"

type Config struct {
	Server struct {
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
	} `yaml:"server"`
}

func loadConfig(yamlPath string) (*Config, error) {
	cfg := &Config{}

	err := readYaml(yamlPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = readEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read env: %w", err)
	}

	return cfg, nil
}

func readYaml(path string, cfg *Config) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)

	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("failed to decode config: %w", err)
	}

	return nil
}

func readEnv(cfg *Config) error {
	err := envconfig.Process("", cfg)
	if err != nil {
		return fmt.Errorf("failed to process env: %w", err)
	}

	return nil
}

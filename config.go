package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config represents the main configuration structure
type Config struct {
	Tailscale TailscaleConfig `json:"tailscale"`
	Services  []ServiceConfig `json:"services"`
}

// TailscaleConfig holds global Tailscale settings
type TailscaleConfig struct {
	AuthKey   string `json:"auth_key"`
	Ephemeral bool   `json:"ephemeral"`
}

// ServiceConfig represents configuration for a single service
type ServiceConfig struct {
	Target   string `json:"target"`
	NodeName string `json:"node_name"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig checks if the configuration is valid
func validateConfig(config *Config) error {
	if config.Tailscale.AuthKey == "" {
		return fmt.Errorf("tailscale auth_key is required")
	}

	if len(config.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	for i, service := range config.Services {
		if service.Target == "" {
			return fmt.Errorf("service[%d]: target is required", i)
		}
		if service.NodeName == "" {
			return fmt.Errorf("service[%d]: node_name is required", i)
		}
	}

	return nil
}

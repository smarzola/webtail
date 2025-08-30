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
	AuthKey       string `json:"auth_key"`
	Ephemeral     bool   `json:"ephemeral"`
	TailnetDomain string `json:"tailnet_domain"`
}

// ServiceConfig represents configuration for a single service
type ServiceConfig struct {
	LocalPort int    `json:"local_port"`
	Hostname  string `json:"hostname"`
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

	if config.Tailscale.TailnetDomain == "" {
		return fmt.Errorf("tailscale tailnet_domain is required")
	}

	if len(config.Services) == 0 {
		return fmt.Errorf("at least one service must be configured")
	}

	for i, service := range config.Services {
		if service.LocalPort <= 0 || service.LocalPort > 65535 {
			return fmt.Errorf("service[%d]: local_port must be between 1 and 65535", i)
		}
		if service.Hostname == "" {
			return fmt.Errorf("service[%d]: hostname is required", i)
		}
	}

	return nil
}

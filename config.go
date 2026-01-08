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
	Docker    DockerConfig    `json:"docker,omitempty"`
}

// TailscaleConfig holds global Tailscale settings
type TailscaleConfig struct {
	AuthKey   string `json:"auth_key"`
	Ephemeral bool   `json:"ephemeral"`
}

// DockerConfig holds Docker client settings
type DockerConfig struct {
	Network    string `json:"network,omitempty"`
	Host       string `json:"host,omitempty"`
	APIVersion string `json:"api_version,omitempty"`
	CertPath   string `json:"cert_path,omitempty"`
	TLSVerify  *bool  `json:"tls_verify,omitempty"`
}

// ServiceConfig represents configuration for a single service
type ServiceConfig struct {
	Target             string `json:"target"`
	NodeName           string `json:"node_name"`
	PassHostHeader     *bool  `json:"pass_host_header,omitempty"`
	TrustForwardHeader *bool  `json:"trust_forward_header,omitempty"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(configPath string, dockerEnabled bool) (*Config, error) {
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

	if err := validateConfig(&config, dockerEnabled); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// boolValue returns the bool value or default if nil
func boolValue(ptr *bool, defaultVal bool) bool {
	if ptr == nil {
		return defaultVal
	}
	return *ptr
}

// validateConfig checks if the configuration is valid
func validateConfig(config *Config, dockerEnabled bool) error {
	if config.Tailscale.AuthKey == "" {
		return fmt.Errorf("tailscale auth_key is required")
	}

	// Services are optional when Docker discovery is enabled
	if len(config.Services) == 0 && !dockerEnabled {
		return fmt.Errorf("at least one service must be configured (or use -docker flag)")
	}

	// Docker network is required when Docker discovery is enabled
	if dockerEnabled && config.Docker.Network == "" {
		return fmt.Errorf("docker.network is required when using -docker flag")
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

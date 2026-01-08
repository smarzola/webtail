package main

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        Config
		dockerEnabled bool
		wantErr       bool
	}{
		{
			name: "valid config with http target",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{
					{
						Target:   "http://localhost:8080",
						NodeName: "test",
					},
				},
			},
			dockerEnabled: false,
			wantErr:       false,
		},
		{
			name: "valid config with https target",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{
					{
						Target:   "https://api.example.com",
						NodeName: "api",
					},
				},
			},
			dockerEnabled: false,
			wantErr:       false,
		},
		{
			name: "invalid config missing target",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{
					{
						NodeName: "test",
					},
				},
			},
			dockerEnabled: false,
			wantErr:       true,
		},
		{
			name: "empty services without docker",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{},
			},
			dockerEnabled: false,
			wantErr:       true,
		},
		{
			name: "empty services with docker enabled but no network",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{},
			},
			dockerEnabled: true,
			wantErr:       true,
		},
		{
			name: "empty services with docker enabled and network",
			config: Config{
				Tailscale: TailscaleConfig{
					AuthKey: "test-key",
				},
				Services: []ServiceConfig{},
				Docker: DockerConfig{
					Network: "webtail",
				},
			},
			dockerEnabled: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config, tt.dockerEnabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

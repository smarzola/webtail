package main

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
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
			wantErr: false,
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
			wantErr: false,
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
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(&tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

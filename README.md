# Webtail - Tailscale Reverse Proxy

A reverse proxy that creates individual Tailscale devices for each service, exposing them on your tailnet with custom hostnames.

## Features

- **Per-service Tailscale devices**: Each service gets its own Tailscale node
- **Automatic hostname assignment**: Services are accessible at `service-name.your-tailnet.ts.net`
- **Ephemeral nodes**: Optional ephemeral node support for temporary deployments
- **HTTP proxying**: Uses oxy library for robust HTTP forwarding
- **Configuration-driven**: All settings managed through a JSON config file
- **Graceful shutdown**: Proper cleanup of all proxy servers

## Prerequisites

- Go 1.19 or later
- A Tailscale account with admin access
- Tailscale auth key (reusable or single-use)

## Installation

1. Clone this repository:
```bash
git clone <repository-url>
cd webtail
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o webtail .
```

## Configuration

Create a `config.json` file in the same directory as the executable:

```json
{
  "tailscale": {
    "auth_key": "tskey-your-auth-key-here",
    "ephemeral": true,
    "tailnet_domain": "your-tailnet.ts.net"
  },
  "services": [
    {
      "name": "plex",
      "local_port": 32400,
      "hostname": "plex"
    },
    {
      "name": "sonarr",
      "local_port": 8989,
      "hostname": "sonarr"
    }
  ]
}
```

### Configuration Fields

#### Tailscale Configuration
- `auth_key`: Your Tailscale auth key (required)
- `ephemeral`: Whether to create ephemeral nodes (optional, default: false)
- `tailnet_domain`: Your tailnet domain (required)

#### Service Configuration
- `name`: Human-readable name for the service (required)
- `local_port`: Port number where the service runs locally (required, 1-65535)
- `hostname`: Hostname for the Tailscale device (required)

## Usage

1. **Configure your services**: Edit `config.json` with your Tailscale credentials and service details.

2. **Start the proxy**:
```bash
./webtail -config config.json
```

3. **Access your services**: Once running, your services will be available at:
   - `plex.your-tailnet.ts.net`
   - `sonarr.your-tailnet.ts.net`

## How It Works

1. **Device Creation**: For each service in the configuration, webtail creates a separate Tailscale node using tsnet.

2. **Hostname Assignment**: Each node gets a hostname based on the service configuration (e.g., `plex` becomes `plex.your-tailnet.ts.net`).

3. **Proxy Setup**: Each node listens on port 80 and forwards requests to the corresponding local service using the oxy HTTP proxy library.

4. **Tailnet Integration**: All nodes automatically join your tailnet and are accessible from any device in your network.

## Security Considerations

- **Auth Keys**: Use Tailscale auth keys with appropriate permissions and expiration
- **Ephemeral Nodes**: Consider using ephemeral nodes for temporary deployments
- **Network Access**: Services are exposed on your tailnet - ensure proper access controls
- **Local Services**: Only expose services that are intended for network access

## Troubleshooting

### Common Issues

1. **"Failed to start tsnet server"**
   - Check your Tailscale auth key is valid
   - Ensure you have network connectivity
   - Verify the auth key has appropriate permissions

2. **"Failed to create listener"**
   - Check if port 80 is available on the system
   - Ensure no other services are using the same port

3. **Services not accessible**
   - Verify the local services are running on the specified ports
   - Check Tailscale node status in your admin console
   - Ensure DNS resolution is working in your tailnet

### Logs

The application provides detailed logging for:
- Proxy startup/shutdown
- Tailscale node creation
- Request forwarding
- Error conditions

## Development

### Building

```bash
go build -o webtail .
```

### Testing

```bash
go test ./...
```

### Dependencies

- `tailscale.com/tsnet`: For creating Tailscale nodes programmatically
- `github.com/vulcand/oxy/forward`: For HTTP proxying and request forwarding

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Support

For issues and questions:
- Check the troubleshooting section above
- Review Tailscale documentation
- Open an issue on the project repository
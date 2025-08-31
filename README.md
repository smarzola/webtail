# Webtail - Tailscale Reverse Proxy

A reverse proxy that creates individual Tailscale devices for each service, exposing them on your tailnet with custom hostnames.

## Features

- **Per-service Tailscale devices**: Each service gets its own Tailscale node
- **Automatic HTTPS certificates**: Tailscale HTTPS provides free SSL certificates
- **Secure access**: Services exposed on port 443 with automatic certificate renewal
- **Automatic hostname assignment**: Services are accessible at `https://service-name.your-tailnet.ts.net`
- **Ephemeral nodes**: Optional ephemeral node support for temporary deployments
- **HTTP proxying**: Uses oxy library for robust HTTP forwarding
- **Configuration-driven**: All settings managed through a JSON config file
- **Graceful shutdown**: Proper cleanup of all proxy servers

## Prerequisites

### For Downloading from Releases:
- A Tailscale account with admin access
- Tailscale auth key (reusable or single-use)
- **Tailscale HTTPS enabled** in your Tailscale admin console

### For Building from Source:
- Go 1.25.0 or later
- A Tailscale account with admin access
- Tailscale auth key (reusable or single-use)
- **Tailscale HTTPS enabled** in your Tailscale admin console

## Enable Tailscale HTTPS

Before using webtail, you need to enable Tailscale HTTPS for automatic certificate management:

1. Go to your [Tailscale Admin Console](https://login.tailscale.com/admin)
2. Navigate to **Settings** → **General**
3. Enable **HTTPS certificates** for your tailnet
4. This allows Tailscale to automatically provision and renew HTTPS certificates for your nodes

## Installation

### Option 1: Download from Releases (Recommended)

1. Go to the [GitHub Releases](https://github.com/smarzola/webtail/releases) page
2. Download the appropriate archive for your platform:
   - `webtail-linux-amd64-vX.X.X.tar.gz` for Linux (AMD64)
   - `webtail-linux-arm64-vX.X.X.tar.gz` for Linux (ARM64)
   - `webtail-darwin-arm64-vX.X.X.tar.gz` for macOS (ARM64)
3. Extract the archive:
   ```bash
   tar -xzf webtail-linux-amd64-v1.0.0.tar.gz
   ```
4. The `webtail` binary is ready to use!

### Option 2: Build from Source

1. Ensure you have Go 1.25.0 or later installed
2. Clone this repository:
   ```bash
   git clone https://github.com/smarzola/webtail.git
   cd webtail
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Build the application:
   ```bash
   go build -o webtail .
   ```

### Using Make (Optional)

If you prefer using Make for building:

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Create release archives
make VERSION=v1.0.0 release-archives
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
      "upstream_host": "localhost:32400",
      "node_name": "plex.your-tailnet.ts.net"
    },
    {
      "upstream_host": "plex-server:32400",
      "node_name": "plex.your-tailnet.ts.net"
    },
    {
      "upstream_host": "192.168.1.100:8989",
      "node_name": "sonarr.your-tailnet.ts.net"
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
- `upstream_host`: Host and port where the upstream service runs (e.g., "localhost:32400", "plex-server:32400", "192.168.1.100:8989") (required)
- `node_name`: Full Tailscale node name (e.g., "plex.your-tailnet.ts.net") (required)

## Usage

1. **Configure your services**: Edit `config.json` with your Tailscale credentials and service details.

2. **Start the proxy**:
```bash
./webtail -config config.json
```

3. **Access your services**: Once running, your services will be available at:
   - `https://plex.your-tailnet.ts.net`
   - `https://sonarr.your-tailnet.ts.net`

   **Note**: Services are exposed on port 443 with automatic HTTPS certificates provided by Tailscale.

## How It Works

1. **Device Creation**: For each service in the configuration, webtail creates a separate Tailscale node using tsnet.

2. **Hostname Assignment**: Each node gets a hostname based on the service configuration (e.g., `plex` becomes `plex.your-tailnet.ts.net`).

3. **Proxy Setup**: Each node listens on port 443 with automatic HTTPS certificates from Tailscale, and forwards requests to the corresponding upstream service using the oxy HTTP proxy library.

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
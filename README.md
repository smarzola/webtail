# Webtail - Tailscale Reverse Proxy

A reverse proxy that creates individual Tailscale nodes for each target service, exposing them on your tailnet with custom hostnames.

## Features

- **Per-service Tailscale nodes**: Each service gets its own Tailscale node
- **Automatic HTTPS certificates**: Tailscale HTTPS provides free SSL certificates
- **Secure access**: Services exposed on port 443 with automatic certificate renewal
- **Automatic hostname assignment**: Services are accessible at `https://service-name.your-tailnet.ts.net`
- **Ephemeral nodes**: Optional ephemeral node support for temporary deployments
- **HTTP proxying**: Uses oxy library for robust HTTP forwarding
- **Configuration-driven**: All settings managed through a JSON config file
- **Docker integration**: Automatic discovery of containers via Docker labels
- **Graceful shutdown**: Proper cleanup of all proxy servers

## Prerequisites

### For Downloading from Releases:
- A Tailscale account with admin access
- Tailscale reusable auth key
- **Tailscale HTTPS enabled** in your Tailscale admin console

### For Building from Source:
- Go 1.25.0 or later
- A Tailscale account with admin access
- Tailscale reusable auth key
- **Tailscale HTTPS enabled** in your Tailscale admin console

## Enable Tailscale HTTPS

Before using webtail, you need to enable Tailscale HTTPS for automatic certificate management:

1. Go to your [Tailscale Admin Console](https://login.tailscale.com/admin/dns)
2. Enable **HTTPS certificates** for your tailnet
3. This allows Tailscale to automatically provision and renew HTTPS certificates for your nodes

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
    "ephemeral": true
  },
  "services": [
    {
      "target": "http://localhost:32400",
      "node_name": "plex"
    },
    {
      "target": "http://plex-server:32400",
      "node_name": "plex-docker",
      "pass_host_header": true
    },
    {
      "target": "http://192.168.1.100:8989",
      "node_name": "sonarr",
      "trust_forward_header": true
    }
  ]
}
```

### Configuration Fields

#### Tailscale Configuration
- `auth_key`: Your Tailscale auth key (required)
- `ephemeral`: Whether to create ephemeral nodes (optional, default: false)

#### Service Configuration
- `target`: Full URL of the upstream service including scheme (e.g., "http://localhost:32400", "https://plex-server:32400", "http://192.168.1.100:8989") (required)
- `node_name`: Tailscale node hostname (e.g., "plex", "api", "web") (required)
- `pass_host_header`: Whether to pass the original Host header to upstream service (optional, default: false)
- `trust_forward_header`: Whether to trust X-Forwarded-* headers from client (optional, default: false)

## Usage

### Configuration-based Mode

1. **Configure your services**: Edit `config.json` with your Tailscale credentials and service details.

2. **Start the proxy**:
```bash
./webtail -config config.json
```

3. **Access your services**: Once running, your services will be available at:
   - `https://plex.your-tailnet.ts.net`
   - `https://sonarr.your-tailnet.ts.net`

   **Note**: The tailnet domain is automatically determined by your Tailscale configuration.

   **Note**: Services are exposed on port 443 with automatic HTTPS certificates provided by Tailscale.

### Docker Discovery Mode

Webtail can automatically discover and proxy Docker containers based on labels. The target URL is built dynamically using the container name and the Docker network specified in the config file.

1. **Configure Docker settings**: Add the `docker` section to your `config.json`:

```json
{
  "tailscale": {
    "auth_key": "tskey-your-auth-key-here",
    "ephemeral": true
  },
  "docker": {
    "network": "webtail"
  },
  "services": []
}
```

With optional Docker client settings:

```json
{
  "tailscale": {
    "auth_key": "tskey-your-auth-key-here",
    "ephemeral": true
  },
  "docker": {
    "network": "webtail",
    "host": "tcp://192.168.1.100:2376",
    "api_version": "1.41",
    "cert_path": "/path/to/certs",
    "tls_verify": true
  },
  "services": []
}
```

2. **Enable Docker mode**:
```bash
./webtail -config config.json -docker
```

3. **Label your containers**: Add the following labels to your Docker containers:

```yaml
# docker-compose.yml example - minimal configuration
services:
  my-app:
    image: my-app:latest
    expose:
      - "8080"
    networks:
      - webtail
    labels:
      webtail.enabled: "true"
      # All other labels are optional:
      # webtail.node_name: "my-app"             # optional, defaults to container name
      # webtail.port: "8080"                    # optional, auto-detected from exposed ports
      # webtail.protocol: "http"                # optional, default: http
      # webtail.pass_host_header: "false"       # optional, default: false
      # webtail.trust_forward_header: "false"   # optional, default: false

networks:
  webtail:
    external: true
```

Or with Docker CLI (minimal):
```bash
docker run -d \
  --name my-app \
  --network webtail \
  --expose 8080 \
  --label webtail.enabled=true \
  my-app:latest
```

With explicit overrides:
```bash
docker run -d \
  --network webtail \
  --label webtail.enabled=true \
  --label webtail.node_name=custom-name \
  --label webtail.port=9000 \
  my-app:latest
```

4. **Automatic lifecycle**: When a labeled container starts, webtail automatically creates a proxy. When the container stops, the proxy is removed.

The target URL is built as: `{protocol}://{container_name}.{docker_network}:{port}`

For example, a container named `my-app` on network `webtail` with port `8080` becomes: `http://my-app.webtail:8080`

#### Docker Labels

| Label | Required | Default | Description |
|-------|----------|---------|-------------|
| `webtail.enabled` | Yes | - | Must be `"true"` to enable proxying |
| `webtail.port` | No | lowest exposed | Container port to proxy to. If not specified, uses the lowest port number among the container's exposed ports |
| `webtail.node_name` | No | container name | Tailscale node hostname. If not specified, uses the container name |
| `webtail.protocol` | No | `http` | Protocol to use (http or https) |
| `webtail.pass_host_header` | No | `false` | Pass original Host header to upstream |
| `webtail.trust_forward_header` | No | `false` | Trust X-Forwarded-* headers from client |

**Port Auto-Detection**: When `webtail.port` is not specified, webtail automatically detects the port by inspecting the container's exposed ports and selecting the lowest port number. For example, if a container exposes ports 80 and 8080, port 80 will be used. If no ports are exposed and no label is provided, the container will be skipped with a warning.

**Node Name Auto-Detection**: When `webtail.node_name` is not specified, the container name is used as the Tailscale node hostname. This allows for minimal configuration - you only need `webtail.enabled=true` if your container has exposed ports.

#### Docker Configuration

The `docker` section in `config.json` supports the following fields:

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `network` | Yes (when using `-docker`) | - | Docker network name for container DNS resolution |
| `host` | No | from env | URL to the Docker server (e.g., `unix:///var/run/docker.sock`, `tcp://localhost:2376`) |
| `api_version` | No | auto | API version to use (leave empty for auto-negotiation) |
| `cert_path` | No | from env | Directory containing TLS certificates (`ca.pem`, `cert.pem`, `key.pem`) |
| `tls_verify` | No | `false` | Enable TLS verification when connecting to Docker |

#### Docker Environment Variables

Alternatively, the Docker client can be configured using environment variables. Environment variables take precedence over config file settings.

| Variable | Description |
|----------|-------------|
| `DOCKER_HOST` | URL to the Docker server (e.g., `unix:///var/run/docker.sock`, `tcp://localhost:2376`) |
| `DOCKER_API_VERSION` | API version to use (leave empty for latest) |
| `DOCKER_CERT_PATH` | Directory containing TLS certificates (`ca.pem`, `cert.pem`, `key.pem`) |
| `DOCKER_TLS_VERIFY` | Enable TLS verification (`1` to enable, empty to disable). Off by default |

Example with remote Docker host using environment variables:
```bash
export DOCKER_HOST=tcp://192.168.1.100:2376
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/path/to/certs
./webtail -config config.json -docker
```

### Combined Mode

You can use both configuration-based and Docker discovery together:

```bash
./webtail -config config.json -docker
```

This allows you to have static services defined in `config.json` alongside dynamic Docker container discovery. When using Docker mode, the `services` array in `config.json` can be empty but `docker.network` is required.

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
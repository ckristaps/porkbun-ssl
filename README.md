# Porkbun SSL Certificate Downloader

A lightweight Go application that downloads SSL certificates from [Porkbun](https://porkbun.com) API. Runs as a long-lived daemon with a built-in cron scheduler for automatic certificate renewal.

## Features

- 🔒 Download SSL certificates from Porkbun API
- 🐳 Docker-ready with multi-stage builds
- ⏰ Built-in cron scheduler for automatic renewal
- 🔐 Secure non-root container execution
- 📁 Customizable certificate paths
- 🔗 Optional combined certificate + private key output

## Prerequisites

- Docker (for containerized deployment) or Go 1.25+ (for local builds)
- Porkbun API credentials ([Get them here](https://porkbun.com/account/api))
- Active SSL certificates on your Porkbun domain

## Quick Start

### 1. Build the Docker Image

```bash
docker build -t porkbun-ssl .
```

### 2. Run the Container

The container runs as a daemon with a built-in scheduler. It downloads certificates immediately on startup and then on the configured schedule.

```bash
docker run -d \
  -e DOMAIN=example.com \
  -e API_KEY=your_api_key \
  -e SECRET_KEY=your_secret_key \
  -v $(pwd)/certs:/certs \
  porkbun-ssl
```

On Windows (PowerShell):
```powershell
docker run -d `
  -e DOMAIN=example.com `
  -e API_KEY=your_api_key `
  -e SECRET_KEY=your_secret_key `
  -v ${PWD}/certs:/certs `
  porkbun-ssl
```

## Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `DOMAIN` | Domain name to download certificates for | `example.com` |
| `API_KEY` | Your Porkbun API key | `pk1_abc123...` |
| `SECRET_KEY` | Your Porkbun secret API key | `sk1_xyz789...` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `API_URL` | Porkbun API endpoint | `https://api.porkbun.com/api/json/v3` |
| `CRON_SCHEDULE` | Cron expression for renewal schedule | `0 2 * * 1` (Mondays at 2 AM) |
| `CERTIFICATE_PATH` | Path template for certificate files | `/certs/{domain}/certificate.pem` |
| `PRIVATE_KEY_PATH` | Path template for private key files | `/certs/{domain}/private_key.pem` |
| `COMBINED_CERT_PATH` | Path for combined cert + key file (if set, separate files are not saved) | `` (disabled) |

**Note:** Paths can contain the `{domain}` placeholder which will be replaced with the domain name.

## Usage Examples

### Basic Usage

```bash
docker run -d \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -v ./certs:/certs \
  porkbun-ssl
```

### Custom Renewal Schedule

```bash
docker run -d \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -e CRON_SCHEDULE="0 3 * * *" \
  -v ./certs:/certs \
  porkbun-ssl
```

Common cron schedules:
- `0 2 * * *` - Daily at 2 AM
- `0 2 * * 1` - Every Monday at 2 AM (default)
- `0 0 1 * *` - First day of each month at midnight

### Custom Certificate Paths

```bash
docker run -d \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -e CERTIFICATE_PATH=/certs/{domain}/fullchain.pem \
  -e PRIVATE_KEY_PATH=/certs/{domain}/privkey.pem \
  -v ./certs:/certs \
  porkbun-ssl
```

### Combined Certificate File

Some applications prefer a single file containing both the certificate chain and private key:

```bash
docker run -d \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -e COMBINED_CERT_PATH=/certs/{domain}/combined.pem \
  -v ./certs:/certs \
  porkbun-ssl
```

## Running with Docker Compose

Create a `.env` file:

```env
DOMAIN=example.com
API_KEY=pk1_abc123
SECRET_KEY=sk1_xyz789
```

Use the provided [docker-compose.yml](docker-compose.yml) and run:

```bash
docker-compose up -d
```

The container will run as a daemon, renewing certificates on the configured schedule.

## Development

### Build Locally

```bash
go build -o porkbun-ssl .
```

### Run Locally

```bash
export DOMAIN=example.com
export API_KEY=your_api_key
export SECRET_KEY=your_secret_key
./porkbun-ssl
```

The application will download certificates immediately and then continue running, renewing on the configured schedule. Press `Ctrl+C` to stop.

### Test

```bash
go test ./...
```

## Output Structure

By default, certificates are saved to:

```
/certs/
└── example.com/
    ├── certificate.pem    # Full certificate chain
    └── private_key.pem    # Private key
```

Or with `COMBINED_CERT_PATH`:

```
/certs/
└── example.com/
    └── combined.pem       # Certificate chain + private key
```

## Integration Examples

### Nginx

Mount certificates and set up a post-renewal hook to reload Nginx:

```yaml
# docker-compose.yml
services:
  porkbun-ssl:
    image: porkbun-ssl:latest
    environment:
      - DOMAIN=example.com
      - API_KEY=pk1_abc123
      - SECRET_KEY=sk1_xyz789
    volumes:
      - /etc/nginx/ssl:/certs
    restart: unless-stopped

  nginx:
    image: nginx:latest
    volumes:
      - /etc/nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - porkbun-ssl
```

### Traefik

Mount certificates to Traefik's certificate directory:

```yaml
services:
  porkbun-ssl:
    image: porkbun-ssl:latest
    environment:
      - DOMAIN=example.com
      - API_KEY=pk1_abc123
      - SECRET_KEY=sk1_xyz789
    volumes:
      - ./certs:/certs
    restart: unless-stopped

  traefik:
    image: traefik:latest
    volumes:
      - ./certs:/certs:ro
```

### HAProxy

HAProxy requires a combined PEM file containing both the certificate chain and private key. Use the `COMBINED_CERT_PATH` option:

```yaml
services:
  porkbun-ssl:
    image: porkbun-ssl:latest
    environment:
      - DOMAIN=example.com
      - API_KEY=pk1_abc123
      - SECRET_KEY=sk1_xyz789
      - COMBINED_CERT_PATH=/certs/{domain}/haproxy.pem
    volumes:
      - ./certs:/certs
    restart: unless-stopped

  haproxy:
    image: haproxy:latest
    volumes:
      - ./certs:/etc/haproxy/certs:ro
    ports:
      - "443:443"
```

Then reference the certificate in your HAProxy configuration:

```haproxy
frontend https_front
    bind *:443 ssl crt /etc/haproxy/certs/example.com/haproxy.pem
    default_backend web_servers
```

## Security Considerations

- Store API credentials securely (use Docker secrets, environment files, or secret managers)
- Never commit `.env` files or credentials to version control
- The container runs as a non-root user (UID 1000)
- Certificate files are created with `0600` permissions (owner read/write only)
- Use HTTPS for all API communications (default)

## Troubleshooting

### "DOMAIN is required but not set"
Make sure you've set the `DOMAIN` environment variable.

### "API error: Authentication failed"
Verify your `API_KEY` and `SECRET_KEY` are correct and have SSL certificate permissions.

### Permission Denied
Ensure the volume mount directory is writable. The container runs as UID 1000.

```bash
mkdir -p certs
chmod 755 certs
```

### Invalid Cron Schedule
Verify your `CRON_SCHEDULE` uses valid cron syntax (5 fields: minute, hour, day of month, month, day of week).

## Similar Projects

- [tmzane/porkcron](https://github.com/tmzane/porkcron) - Python-based Porkbun certificate downloader

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## API Documentation

For more information about the Porkbun API, see:
- [Porkbun API Documentation](https://porkbun.com/api/json/v3/documentation)

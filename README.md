# Porkbun SSL Certificate Downloader

A lightweight Go application that downloads SSL certificates from [Porkbun](https://porkbun.com) API. Designed to run as a Docker container and can be scheduled with cron for automatic certificate renewal.

## Features

- 🔒 Download SSL certificates from Porkbun API
- 🐳 Docker-ready with multi-stage builds
- 📦 Supports multiple domains in a single run
- ⏰ Easy to schedule with cron
- 🔐 Secure non-root container execution
- 📁 Customizable certificate paths

## Prerequisites

- Docker (for containerized deployment)
- Porkbun API credentials ([Get them here](https://porkbun.com/account/api))
- Active SSL certificates on your Porkbun domains

## Quick Start

### 1. Build the Docker Image

```bash
docker build -t porkbun-ssl .
```

### 2. Run the Container

```bash
docker run --rm \
  -e DOMAIN=example.com \
  -e API_KEY=your_api_key \
  -e SECRET_KEY=your_secret_key \
  -v $(pwd)/certs:/certs \
  porkbun-ssl
```

On Windows (PowerShell):
```powershell
docker run --rm `
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
| `DOMAIN` | Domain name(s) to download certificates for (comma-separated) | `example.com,example.org` |
| `API_KEY` | Your Porkbun API key | `pk1_abc123...` |
| `SECRET_KEY` | Your Porkbun secret API key | `sk1_xyz789...` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `API_URL` | Porkbun API endpoint | `https://api.porkbun.com/api/json/v3` |
| `CERTIFICATE_PATH` | Path template for certificate files | `/certs/{domain}/certificate.pem` |
| `PRIVATE_KEY_PATH` | Path template for private key files | `/certs/{domain}/private_key.pem` |

**Note:** When downloading certificates for multiple domains, paths must contain the `{domain}` placeholder.

## Usage Examples

### Single Domain

```bash
docker run --rm \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -v ./certs:/certs \
  porkbun-ssl
```

### Multiple Domains

```bash
docker run --rm \
  -e DOMAIN=example.com,blog.example.com,api.example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -v ./certs:/certs \
  porkbun-ssl
```

### Custom Certificate Paths

```bash
docker run --rm \
  -e DOMAIN=example.com \
  -e API_KEY=pk1_abc123 \
  -e SECRET_KEY=sk1_xyz789 \
  -e CERTIFICATE_PATH=/certs/{domain}/fullchain.pem \
  -e PRIVATE_KEY_PATH=/certs/{domain}/privkey.pem \
  -v ./certs:/certs \
  porkbun-ssl
```

## Scheduling with Cron

### Using Docker Compose

Create a `.env` file:

```env
DOMAIN=example.com,example.org
API_KEY=pk1_abc123
SECRET_KEY=sk1_xyz789
```

Use the provided [docker-compose.yml](docker-compose.yml) and run:

```bash
docker-compose up -d
```

### Using Host Cron

Add to your crontab (`crontab -e`):

```bash
# Run daily at 2 AM
0 2 * * * docker run --rm --env-file /path/to/.env -v /certs:/certs porkbun-ssl:latest >> /var/log/porkbun-ssl.log 2>&1
```

### Using Kubernetes CronJob

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: porkbun-ssl
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: porkbun-ssl
            image: porkbun-ssl:latest
            env:
            - name: DOMAIN
              valueFrom:
                secretKeyRef:
                  name: porkbun-credentials
                  key: domain
            - name: API_KEY
              valueFrom:
                secretKeyRef:
                  name: porkbun-credentials
                  key: api-key
            - name: SECRET_KEY
              valueFrom:
                secretKeyRef:
                  name: porkbun-credentials
                  key: secret-key
            volumeMounts:
            - name: certs
              mountPath: /certs
          volumes:
          - name: certs
            persistentVolumeClaim:
              claimName: ssl-certs
          restartPolicy: OnFailure
```

## Development

### Build Locally

```bash
go build -o porkbun-ssl app.go
```

### Run Locally

```bash
export DOMAIN=example.com
export API_KEY=your_api_key
export SECRET_KEY=your_secret_key
./porkbun-ssl
```

### Test

```bash
go test ./...
```

## Output Structure

By default, certificates are saved to:

```
/certs/
├── example.com/
│   ├── certificate.pem    # Full certificate chain
│   └── private_key.pem    # Private key
└── example.org/
    ├── certificate.pem
    └── private_key.pem
```

## Integration Examples

### Nginx

After downloading certificates, reload Nginx:

```bash
docker run --rm \
  --env-file .env \
  -v /etc/nginx/ssl:/certs \
  porkbun-ssl && \
nginx -s reload
```

### Traefik

Mount certificates to Traefik's certificate directory:

```bash
docker run --rm \
  --env-file .env \
  -v /etc/traefik/certs:/certs \
  porkbun-ssl
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

### Multiple Domains Without Placeholder
When using multiple domains, ensure paths contain `{domain}`:

```bash
-e CERTIFICATE_PATH=/certs/{domain}/cert.pem
```

## Similar Projects

- [tmzane/porkcron](https://github.com/tmzane/porkcron) - Python-based Porkbun certificate downloader

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## API Documentation

For more information about the Porkbun API, see:
- [Porkbun API Documentation](https://porkbun.com/api/json/v3/documentation)

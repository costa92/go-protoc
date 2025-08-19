# Grafana Configuration Files

This directory contains configuration templates for Grafana installation.

## Files Overview

### Configuration Files
- `grafana.ini` - Host/native installation comprehensive configuration template
- `grafana-docker.ini` - Docker container installation simplified configuration template

### Service Files
- `grafana.service` - systemd service file template for native installation
- `grafana-server` - Environment variables file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_GRAFANA_HOST` - Grafana server host (default: 127.0.0.1)
- `PROJ_GRAFANA_PORT` - Grafana server port (default: 3000)
- `PROJ_GRAFANA_ADMIN_USER` - Admin username (default: admin)
- `PROJ_GRAFANA_ADMIN_PASSWORD` - Admin password (default: proj(#)666)

### Directories
- `PROJ_GRAFANA_DATA_DIR` - Data directory (default: /var/lib/grafana)

### Integration (Docker networking)
- `NETWORK_NAME` - Docker network name for container service discovery
- `PROJ_JAEGER_HOST` - Jaeger host for distributed tracing
- `PROJ_JAEGER_AGENT_COMPACT_PORT` - Jaeger agent UDP port

### System
- `HOSTNAME` - System hostname for instance identification

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. Web server starts on configured host and port

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container
3. Container uses service discovery via Docker network
4. Admin credentials are set via environment variables

## Configuration Types

### Host Installation (grafana.ini)
- Full Grafana configuration with all sections
- Uses host-specific paths and networking
- Includes security hardening options
- Supports file-based logging and data storage
- Jaeger tracing points to host-based Jaeger instance

### Docker Installation (grafana-docker.ini)
- Simplified configuration optimized for containers
- Uses container internal paths (/var/lib/grafana)
- Console logging for Docker log aggregation  
- Redis cache can use Docker service discovery
- Jaeger tracing uses container network names

## Key Features Configured

### Authentication & Security
- Admin user/password configuration
- Session security settings  
- HTTPS/security headers support
- Anonymous access disabled by default

### Database
- SQLite3 by default (suitable for development)
- MySQL/PostgreSQL support available
- Connection pooling configured

### Monitoring & Observability
- Internal metrics enabled (/metrics endpoint)
- Jaeger distributed tracing integration
- Alerting system enabled
- Log aggregation configured

### UI & UX
- Dark theme by default
- Explore section enabled
- Dashboard versioning (20 versions kept)
- Plugin management disabled by default

## Integration Points

### With Prometheus
- Metrics endpoint available at `/metrics`
- Can be scraped by Prometheus configuration

### With Jaeger
- Distributed tracing enabled
- Spans sent to configured Jaeger agent
- Sampling configured for development (100%)

### With Docker Services
- Service discovery via Docker networking
- Can access other containerized services
- Redis caching can use container names

## Security Considerations

- Admin password should be changed after first login
- Secret key is configured (should be unique in production)
- Anonymous access disabled
- User registration disabled
- Plugin installation disabled by default

## Directory Structure

After installation:
```
/var/lib/grafana/          # Data directory
├── grafana.db            # SQLite database
├── plugins/              # Plugin directory
└── ...

/var/log/grafana/         # Log directory
├── grafana.log           # Main log file
└── ...

/etc/grafana/             # Configuration directory  
├── grafana.ini           # Main configuration
├── provisioning/         # Auto-provisioning configs
└── ...
```
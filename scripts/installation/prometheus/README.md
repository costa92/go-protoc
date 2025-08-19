# Prometheus Configuration Files

This directory contains configuration templates for Prometheus installation.

## Files Overview

### Configuration Files
- `prometheus.yml` - Host/native installation configuration template
- `prometheus-docker.yml` - Docker container installation configuration template

### Service Files
- `prometheus.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Host Configuration
- `PROJ_PROMETHEUS_HOST` - Prometheus server host (default: 127.0.0.1)
- `PROJ_PROMETHEUS_PORT` - Prometheus server port (default: 9090)

### Directories
- `PROJ_PROMETHEUS_CONFIG_DIR` - Configuration directory (default: /etc/prometheus)
- `PROJ_PROMETHEUS_DATA_DIR` - Data directory (default: /var/lib/prometheus)

### Other Services (for scraping)
- `PROJ_JAEGER_HOST` - Jaeger host for metrics scraping
- `PROJ_JAEGER_COLLECTOR_HTTP_PORT` - Jaeger collector HTTP port
- `PROJ_VICTORIAMETRICS_HOST` - VictoriaMetrics host
- `PROJ_VICTORIAMETRICS_PORT` - VictoriaMetrics port
- `PROJ_ALERTMANAGER_HOST` - AlertManager host
- `PROJ_ALERTMANAGER_PORT` - AlertManager port

### Network Configuration
- `NETWORK_NAME` - Docker network name (for container service discovery)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container
3. Container uses service discovery via Docker network

## Scrape Targets

### Host Installation (prometheus.yml)
- Prometheus itself (self-monitoring)
- go-protoc API service
- Node exporter (system metrics)
- Jaeger metrics
- VictoriaMetrics
- Redis exporter

### Docker Installation (prometheus-docker.yml)
- All containerized services using Docker service discovery
- Host-based services via `host.docker.internal`

## Storage Configuration

- **Retention Time**: 30 days
- **Retention Size**: 10GB
- **TSDB Path**: Configured per environment

## Security

The systemd service includes security hardening:
- NoNewPrivileges
- PrivateTmp
- ProtectSystem=strict
- ProtectHome
- Limited file access

## Monitoring Features

- **Web Console**: Access via http://host:9090
- **Metrics API**: Standard Prometheus API endpoints
- **Lifecycle Management**: Web lifecycle endpoints enabled
- **Auto-discovery**: Service discovery for containerized environments
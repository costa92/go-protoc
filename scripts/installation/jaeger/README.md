# Jaeger Configuration Files

This directory contains configuration templates for Jaeger installation.

## Files Overview

### Configuration Files
- `jaeger-config.yaml` - Host/native installation configuration template
- `jaeger-config-docker.yaml` - Docker container installation configuration template

### Service Files
- `jaeger.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Host Configuration
- `PROJ_JAEGER_HOST` - Jaeger server host (default: 127.0.0.1)
- `PROJ_JAEGER_QUERY_PORT` - Query service port (default: 16686)
- `PROJ_JAEGER_COLLECTOR_HTTP_PORT` - Collector HTTP port (default: 14268)
- `PROJ_JAEGER_COLLECTOR_GRPC_PORT` - Collector gRPC port (default: 14250)
- `PROJ_JAEGER_AGENT_HTTP_PORT` - Agent HTTP port (default: 5778)
- `PROJ_JAEGER_AGENT_COMPACT_PORT` - Agent compact thrift port (default: 6831)
- `PROJ_JAEGER_AGENT_BINARY_PORT` - Agent binary thrift port (default: 6832)

### Directories
- `PROJ_JAEGER_CONFIG_DIR` - Configuration directory (default: /etc/jaeger)
- `PROJ_JAEGER_DATA_DIR` - Data directory (default: /var/lib/jaeger)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container
3. Container ports are mapped to host

## Configuration Types

### Host Installation (jaeger-config.yaml)
- Uses host IP addresses and ports
- Suitable for native/binary installation
- Integrates with systemd

### Docker Installation (jaeger-config-docker.yaml)
- Uses container internal addresses (0.0.0.0)
- Port mapping handled by Docker
- Optimized for containerized deployment

## Security

The systemd service includes security hardening:
- NoNewPrivileges
- PrivateTmp
- ProtectSystem=strict
- ProtectHome
- Limited file access

## Storage Options

Currently configured for memory storage (development/testing). For production:
- Elasticsearch: Recommended for large deployments
- Cassandra: Good for high-volume scenarios  
- Badger: Embedded storage option
# AlertManager Configuration Files

This directory contains configuration templates for AlertManager installation.

## Files Overview

### Configuration Files
- `alertmanager.yml` - Host/native installation configuration template
- `alertmanager-docker.yml` - Docker container installation configuration template

### Service Files
- `alertmanager.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Host Configuration
- `PROJ_ALERTMANAGER_HOST` - AlertManager server host (default: 127.0.0.1)
- `PROJ_ALERTMANAGER_PORT` - AlertManager server port (default: 9093)

### Directories
- `PROJ_ALERTMANAGER_CONFIG_DIR` - Configuration directory (default: /etc/alertmanager)
- `PROJ_ALERTMANAGER_DATA_DIR` - Data directory (default: /var/lib/alertmanager)

### Docker Networking
- `NETWORK_NAME` - Docker network name for container service discovery

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. Web UI available on configured host and port

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container
3. Container uses service discovery via Docker network
4. Webhook URLs point to host via `host.docker.internal`

## Configuration Features

### Routing and Grouping
- **Group By**: Alerts are grouped by `alertname`, `cluster`, and `service`
- **Timing**: 10s group wait, 10s group interval, 1h repeat interval
- **Routes**: Separate handling for databases, containers, and critical alerts

### Alert Classifications

#### Host Installation (alertmanager.yml)
- **Database Alerts**: MySQL, Cassandra services
- **Development Alerts**: Environment-specific routing
- **Critical Alerts**: Immediate notification with reduced intervals
- **General Alerts**: Default webhook handling

#### Docker Installation (alertmanager-docker.yml)
- **Database Alerts**: MySQL, MariaDB, MongoDB, Redis containers
- **Container Alerts**: Docker container-specific alerts
- **Critical Alerts**: High-priority immediate notifications
- **High Priority Alerts**: Fast-tracked important alerts

### Inhibition Rules
- Critical alerts suppress warning and info level alerts
- Container down alerts inhibit related service alerts
- Prevents alert spam during cascading failures

### Receivers and Notifications

#### Webhook Receivers
- **Default Webhook**: General purpose alert handling
- **Database Webhook**: Database-specific alert formatting
- **Container Webhook**: Container-aware alert information
- **Critical Webhook**: Enhanced formatting for urgent alerts

#### Notification Channels (Examples Provided)
- **Email**: SMTP-based email notifications
- **Slack**: Slack webhook integration with rich formatting
- **Custom Webhooks**: Flexible integration endpoints

### Time Intervals
- **Business Hours**: Monday-Friday, 9 AM-5 PM (host)
- **Dev Hours**: Extended hours for development (container)
- **Maintenance Windows**: Scheduled maintenance periods
- **Weekend**: Weekend-specific handling

## Integration Points

### With Prometheus
AlertManager receives alerts from Prometheus via webhook or API calls.

### With Docker Services
- Container-aware alert routing
- Network-based service discovery
- Host access via `host.docker.internal`

### With Monitoring Stack
- Jaeger tracing integration available
- Metrics exposure for monitoring AlertManager itself
- Log aggregation compatibility

## Alert Flow

```
Prometheus → [Rules] → AlertManager → [Routes] → Receivers → Notifications
     ↓              ↓                    ↓           ↓            ↓
  Metrics      Alert Rules         Grouping    Webhooks    External Systems
```

### Example Alert Routing

1. **Critical Database Alert**:
   - Route: `service: mysql` + `severity: critical`
   - Receiver: `critical-alerts` (immediate)
   - Notification: Webhook with 🚨 formatting

2. **Container Down Alert**:
   - Route: `container: proj-redis`
   - Receiver: `container-alerts`
   - Inhibits: Other redis service alerts
   - Notification: Container-specific webhook

3. **General Warning**:
   - Route: Default route
   - Receiver: `web.hook`
   - Notification: Standard webhook format

## Security and Reliability

### Security Features
- systemd security hardening (NoNewPrivileges, PrivateTmp)
- Restricted file system access
- Limited network exposure

### Reliability Features
- Automatic service restart on failure
- Persistent alert state storage
- Graceful shutdown with SIGTERM handling
- File descriptor limits configured

## Customization

### Adding New Receivers
1. Add receiver configuration in the `receivers:` section
2. Create corresponding route in `routes:` section
3. Configure inhibition rules if needed
4. Test with sample alerts

### Webhook Payload Format
AlertManager sends JSON payloads with alert information:
```json
{
  "receiver": "web.hook",
  "status": "firing",
  "alerts": [...],
  "groupLabels": {...},
  "commonLabels": {...},
  "commonAnnotations": {...}
}
```

## Troubleshooting

### Common Issues
1. **Webhook timeouts**: Check target service availability
2. **Alert not routing**: Verify label matching in routes
3. **Duplicate notifications**: Review inhibition rules
4. **SMTP failures**: Validate SMTP configuration

### Debugging
- Check AlertManager logs: `journalctl -f -u alertmanager`
- Web UI: `http://host:9093`
- API endpoint: `http://host:9093/api/v1/alerts`
- Configuration validation: `amtool config check alertmanager.yml`
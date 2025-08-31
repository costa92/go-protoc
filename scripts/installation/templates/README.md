# Multi-Platform Configuration Templates

This directory contains configuration templates for all supported services across different deployment platforms.

## Directory Structure

```
templates/
├── docker/                  # Docker deployment templates
│   ├── redis/
│   │   ├── docker-compose.yml.tpl
│   │   └── redis.conf.tpl
│   ├── otelcol/
│   │   ├── docker-compose.yml.tpl
│   │   └── config.yaml.tpl
│   └── victorialogs/
│       └── docker-compose.yml.tpl
├── ubuntu/                  # Ubuntu native deployment templates
│   ├── redis/
│   │   ├── systemd.service.tpl
│   │   └── redis.conf.tpl
│   └── ...
├── macos/                   # macOS native deployment templates
│   ├── redis/
│   │   ├── launchd.plist.tpl
│   │   └── redis.conf.tpl
│   └── ...
└── README.md               # This file
```

## Template Variables

All templates use environment variable substitution with the format `${VARIABLE_NAME}` or `${VARIABLE_NAME:-default_value}`.

### Global Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PROJ_NAME` | Project name | `go-protoc` |
| `PROJ_ENVIRONMENT` | Deployment environment | `development` |
| `PROJ_ROOT_DIR` | Project root directory | Auto-detected |
| `PROJ_SERVICE_NAME` | Current service name | `apiserver` |
| `PROJ_SERVICE_VERSION` | Current service version | `v2.0.0` |

### Service-Specific Variables

Each service has its own set of variables defined in `scripts/installation/versions.sh`:

#### Redis
- `REDIS_VERSION` - Redis version
- `PROJ_REDIS_PORT` - Redis port (default: 6379)
- `PROJ_REDIS_CONFIG_DIR` - Config directory
- `PROJ_REDIS_DATA_DIR` - Data directory
- `REDIS_MAX_MEMORY` - Memory limit (default: 256mb)
- `REDIS_PASSWORD` - Authentication password (optional)

#### OpenTelemetry Collector
- `OTELCOL_VERSION` - OTEL Collector version
- `PROJ_OTELCOL_GRPC_PORT` - OTLP gRPC port (default: 4327)
- `PROJ_OTELCOL_HTTP_PORT` - OTLP HTTP port (default: 4328)
- `PROJ_OTELCOL_CONFIG_DIR` - Config directory
- `PROJ_OTELCOL_DATA_DIR` - Data directory

#### VictoriaLogs
- `VICTORIALOGS_VERSION` - VictoriaLogs version
- `PROJ_VICTORIALOGS_PORT` - VictoriaLogs port (default: 9428)
- `PROJ_VICTORIALOGS_CONFIG_DIR` - Config directory
- `PROJ_VICTORIALOGS_DATA_DIR` - Data directory
- `VICTORIALOGS_RETENTION` - Log retention period (default: 7d)

## Template Management

### Generate Configuration Files

Use the template manager to generate actual configuration files:

```bash
# Load the template manager library
source scripts/installation/lib/template_manager.sh

# Generate a single config file
proj::template::generate_config "redis" "docker" "docker-compose.yml.tpl" "/path/to/output/docker-compose.yml"

# Generate all config files for a service
proj::template::generate_service_configs "redis" "docker" "/path/to/config/dir"
```

### List Available Templates

```bash
# List all platforms
./scripts/installation/lib/template_manager.sh

# List services for a platform
./scripts/installation/lib/template_manager.sh docker

# List templates for a specific service
./scripts/installation/lib/template_manager.sh docker redis
```

### Validate Templates

```bash
# Validate a specific template
proj::template::validate_template "templates/docker/redis/redis.conf.tpl"
```

### Create New Templates

```bash
# Create a new template
proj::template::create_template "myservice" "docker" "config.yml.tpl"
```

## Platform-Specific Details

### Docker Templates

Docker templates focus on:
- Pure docker command orchestration (no docker-compose dependency)
- Volume mounting for persistent data
- Network configuration for service communication
- Health checks and restart policies
- Environment variable injection
- Script-based container lifecycle management

Key files:
- `docker-run.sh.tpl` - Container startup script
- `docker-stop.sh.tpl` - Container shutdown script
- `docker-status.sh.tpl` - Container status check script
- Service-specific config files (e.g., `redis.conf.tpl`)

### Ubuntu Templates

Ubuntu templates focus on:
- systemd service definitions
- APT package integration
- System user and directory management
- Security settings and resource limits
- Log management with journald

Key files:
- `systemd.service.tpl` - systemd service definition
- Service-specific config files optimized for system integration

### macOS Templates

macOS templates focus on:
- launchd service definitions
- Homebrew package integration
- User-space service management
- macOS-specific file paths and permissions
- Integration with system logging

Key files:
- `launchd.plist.tpl` - launchd service definition
- Service-specific config files optimized for macOS

## Template Conventions

### Naming Conventions

- Template files must end with `.tpl` extension
- Generated files will have the `.tpl` extension removed
- Use descriptive names that indicate the file purpose

### Variable Naming

- Use `PROJ_` prefix for project-specific variables
- Use service name prefix for service-specific variables (e.g., `REDIS_`, `OTELCOL_`)
- Use UPPER_CASE for all environment variables
- Provide sensible defaults using `${VAR:-default}` syntax

### Comments and Documentation

- Include service name, version, and purpose in template headers
- Document any non-obvious configuration choices
- Include references to official documentation where applicable

### Security Considerations

- Never include hardcoded secrets or passwords
- Use environment variables for sensitive configuration
- Set appropriate file permissions in generated configs
- Follow principle of least privilege for service users

## Integration with Installation Scripts

Templates are automatically used by the installation scripts:

1. **Detection Phase**: Platform and installation method are detected
2. **Template Selection**: Appropriate templates are selected based on platform
3. **Variable Loading**: Environment variables are loaded from `versions.sh`
4. **Generation Phase**: Configuration files are generated from templates
5. **Installation Phase**: Generated configs are used for service installation

This ensures consistent configuration across all deployment methods while maintaining platform-specific optimizations.

## Troubleshooting

### Common Issues

1. **Missing Variables**: Ensure all required environment variables are set
2. **Template Not Found**: Check that the template exists for the specified platform/service
3. **Permission Errors**: Ensure write permissions to the output directory
4. **Variable Substitution Errors**: Validate template syntax with `proj::template::validate_template`

### Debug Mode

Enable debug logging to see detailed template processing:

```bash
export PROJ_LOG_LEVEL=debug
./scripts/installation/lib/template_manager.sh docker redis
```

This will show all variables being substituted and the template processing steps.
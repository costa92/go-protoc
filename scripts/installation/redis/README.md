# Redis Configuration Files

This directory contains configuration templates for Redis installation.

## Files Overview

### Configuration Files
- `redis.conf` - Host/native installation comprehensive configuration template
- `redis-docker.conf` - Docker container installation optimized configuration template

### Service Files
- `redis.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_REDIS_HOST` - Redis server host (default: 127.0.0.1)
- `PROJ_REDIS_PORT` - Redis server port (default: 6379)

### Directories
- `PROJ_REDIS_DATA_DIR` - Data directory (default: /var/lib/redis)
- `PROJ_REDIS_CONFIG_DIR` - Configuration directory (default: /etc/redis)

### Security (Optional)
- `REDIS_PASSWORD` - Redis authentication password (commented out by default)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. Redis runs as `redis` user with proper permissions

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container at `/usr/local/etc/redis/redis.conf`
3. Data persisted to host directory via volume mount
4. Runs with optimized settings for containerized environment

## Configuration Types

### Host Installation (redis.conf)
- **Comprehensive Configuration**: Full Redis configuration with all options
- **System Integration**: Proper daemon mode, logging to files, PID management
- **Security Hardening**: Bind to specific interface, proper permissions
- **Persistence**: Both RDB snapshots and AOF for durability
- **Performance**: Optimized for production server environment

### Docker Installation (redis-docker.conf)
- **Container Optimized**: Adapted for containerized environment
- **Simplified Logging**: Console output for Docker log aggregation
- **Data Persistence**: Volume-mounted data directory (/data)
- **Network**: Binds to all interfaces (0.0.0.0) for container access
- **Resource Limits**: Conservative settings suitable for development

## Key Features Configured

### Persistence
- **RDB Snapshots**: Automatic database snapshots (save points)
- **AOF (Append Only File)**: Transaction log for point-in-time recovery
- **Hybrid Persistence**: RDB + AOF for optimal durability/performance

### Memory Management
- **Eviction Policy**: `allkeys-lru` for automatic memory management
- **Memory Limits**: Configurable max memory usage
- **Lazy Freeing**: Non-blocking memory deallocation

### Security
- **Authentication**: Optional password protection (commented out)
- **Access Control**: Bind address restrictions
- **Command Renaming**: Ability to rename dangerous commands

### Performance
- **I/O Threading**: Multi-threaded I/O support for high throughput
- **Active Defragmentation**: Memory fragmentation reduction
- **Kernel Optimizations**: THP disable, OOM score adjustment

### Monitoring
- **Slow Log**: Track slow-running commands
- **Latency Monitoring**: Track latency sources
- **Client Limits**: Connection and buffer limits

## Integration Points

### With Applications
- Standard Redis protocol on configured port
- Optional authentication if password enabled
- Multiple database support (16 databases by default)

### With Monitoring Stack
- **Prometheus**: Redis metrics can be exported via redis_exporter
- **Logging**: Structured logs for aggregation
- **Health Checks**: Built-in PING command for health monitoring

### With Backup Systems
- **RDB Files**: Point-in-time database snapshots
- **AOF Files**: Incremental transaction logs
- **Replication**: Master-slave replication support

## Directory Structure

After installation:

### Host Installation
```
/var/lib/redis/           # Data directory
├── dump.rdb             # Database snapshot
├── appendonly.aof       # Transaction log
└── ...

/var/log/redis/          # Log directory
├── redis-server.log     # Main log file
└── ...

/etc/redis/              # Configuration directory
├── redis.conf           # Main configuration
└── ...

/run/redis/              # Runtime directory
├── redis-server.pid     # Process ID file
└── ...
```

### Docker Installation
```
/data/                   # Container data directory
├── dump.rdb            # Database snapshot
├── appendonly.aof      # Transaction log
└── ...
```

## Security Considerations

### systemd Service Security
- **NoNewPrivileges**: Prevents privilege escalation
- **PrivateTmp**: Private /tmp directory
- **ProtectSystem**: Read-only system directories
- **Resource Limits**: File descriptor and memory limits

### Network Security
- **Bind Address**: Restrict to specific interfaces
- **Authentication**: Optional password protection
- **Command Filtering**: Rename or disable dangerous commands

### File System Security
- **User Isolation**: Runs as dedicated `redis` user
- **Directory Permissions**: Restricted access to data directories
- **Log Security**: Secure log file permissions

## Performance Tuning

### Memory Settings
- **maxmemory**: Set appropriate memory limits
- **maxmemory-policy**: Choose eviction strategy
- **Memory sampling**: Configure LRU sampling accuracy

### Persistence Settings
- **Save frequency**: Adjust RDB save intervals
- **AOF sync**: Configure fsync frequency
- **Background saves**: Optimize I/O scheduling

### Network Settings
- **TCP backlog**: Adjust connection queue size
- **Timeout**: Configure client idle timeouts
- **Keepalive**: TCP keepalive settings

## Troubleshooting

### Common Issues
1. **Permission denied**: Check file/directory permissions
2. **Cannot bind**: Port already in use or permission issues
3. **Memory issues**: Check memory limits and eviction policy
4. **Slow performance**: Review slow log and configuration

### Debugging
- **Redis logs**: Check `/var/log/redis/redis-server.log`
- **systemd status**: `systemctl status redis`
- **Redis CLI**: `redis-cli ping` for connectivity test
- **Monitor command**: `redis-cli monitor` for real-time commands

### Monitoring Commands
```bash
# Check Redis status
redis-cli ping

# Get server info
redis-cli info

# Monitor slow queries
redis-cli slowlog get 10

# Check memory usage
redis-cli info memory

# Monitor real-time commands
redis-cli monitor
```
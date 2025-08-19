# MariaDB Configuration Files

This directory contains configuration templates for MariaDB installation.

## Files Overview

### Configuration Files
- `50-server.cnf` - Host/native installation server configuration template
- `mariadb-docker.cnf` - Docker container installation optimized configuration template
- `mariadb-client.cnf` - Client configuration template

### Service Files
- `mariadb.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_MYSQL_HOST` - MariaDB server host (default: 127.0.0.1)
- `PROJ_MYSQL_PORT` - MariaDB server port (default: 3306)

### Authentication
- `PROJ_MYSQL_ADMIN_USERNAME` - Root/admin username (default: root)
- `PROJ_MYSQL_ADMIN_PASSWORD` - Root/admin password
- `PROJ_PASSWORD` - Legacy password variable

### Directories
- `PROJ_MYSQL_DATA_DIR` - Data directory (default: /var/lib/mysql)
- `PROJ_MYSQL_CONFIG_DIR` - Configuration directory (default: /etc/mysql)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. MariaDB runs as `mysql` user with proper permissions

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container at `/etc/mysql/conf.d/`
3. Data persisted to host directory via volume mount
4. Runs with optimized settings for containerized environment

## Configuration Types

### Host Installation (50-server.cnf)
- **Production Ready**: Full MariaDB server configuration
- **System Integration**: Proper daemon mode, logging to files, PID management
- **Security Hardening**: Bind to specific interface, proper permissions
- **Performance**: Optimized for production server environment
- **InnoDB Engine**: Full InnoDB configuration with buffer pool tuning

### Docker Installation (mariadb-docker.cnf)
- **Container Optimized**: Adapted for containerized environment
- **Simplified Logging**: Console output for Docker log aggregation
- **Data Persistence**: Volume-mounted data directory (/var/lib/mysql)
- **Network**: Binds to all interfaces (0.0.0.0) for container access
- **Resource Limits**: Conservative settings suitable for development

### Client Configuration (mariadb-client.cnf)
- **Connection Defaults**: Standard client connection settings
- **Character Set**: UTF-8 support for international characters
- **SSL Settings**: Optional SSL configuration for secure connections

## Key Features Configured

### Performance Tuning
- **InnoDB Buffer Pool**: Optimized memory allocation for caching
- **Query Cache**: Improved query performance through caching
- **Connection Limits**: Appropriate connection pooling
- **Thread Handling**: Efficient thread management

### Security
- **Network Binding**: Restrict to specific interfaces
- **Authentication**: Secure password handling
- **SSL Support**: Optional encrypted connections
- **User Management**: Proper user privilege separation

### Reliability
- **Binary Logging**: Transaction logging for replication
- **Error Logging**: Comprehensive error tracking
- **Slow Query Log**: Performance monitoring
- **Crash Recovery**: InnoDB crash recovery settings

### Storage Engine
- **InnoDB Optimization**: Memory, I/O, and locking optimizations
- **File Per Table**: Separate tablespace files
- **Auto Increment**: Proper auto increment locking
- **Foreign Keys**: Full referential integrity support

## Integration Points

### With Applications
- Standard MySQL protocol on configured port
- UTF-8 character set support for international data
- Multiple database support with proper isolation

### With Monitoring Stack
- **Prometheus**: MariaDB metrics can be exported via mysqld_exporter
- **Logging**: Structured logs for aggregation
- **Health Checks**: Built-in status tables for monitoring

### With Backup Systems
- **Binary Logs**: Point-in-time recovery support
- **InnoDB**: Hot backup capabilities
- **Replication**: Master-slave replication ready

## Directory Structure

After installation:

### Host Installation
```
/var/lib/mysql/              # Data directory
├── mysql/                   # System database
├── performance_schema/      # Performance monitoring
├── information_schema/      # Metadata database
└── [user_databases]/        # Application databases

/var/log/mysql/              # Log directory
├── error.log               # Error log
├── mysql-slow.log          # Slow query log
└── mysql-bin.000001        # Binary logs

/etc/mysql/                  # Configuration directory
├── my.cnf                  # Main configuration
├── mariadb.conf.d/         # MariaDB specific configs
│   └── 50-server.cnf       # Server configuration
└── conf.d/                 # Additional configurations
    └── mysql.cnf           # MySQL compatibility
```

### Docker Installation
```
/var/lib/mysql/              # Container data directory
├── mysql/                   # System database
├── performance_schema/      # Performance monitoring
└── [user_databases]/        # Application databases

/etc/mysql/conf.d/           # Container config directory
└── mariadb-docker.cnf      # Server configuration
```

## Security Considerations

### systemd Service Security
- **User Isolation**: Runs as dedicated `mysql` user
- **File System Security**: Restricted access to data directories
- **Resource Limits**: Memory and connection limits
- **Process Isolation**: Proper process containment

### Network Security
- **Bind Address**: Restrict to specific interfaces
- **Authentication**: Strong password requirements
- **SSL Support**: Optional encrypted connections
- **User Privileges**: Principle of least privilege

### File System Security
- **Directory Permissions**: Secure data directory access
- **Log Security**: Proper log file permissions
- **Configuration Protection**: Secure config file access

## Performance Tuning

### Memory Settings
- **innodb_buffer_pool_size**: Primary cache for InnoDB data
- **query_cache_size**: Query result caching
- **sort_buffer_size**: Sorting operations memory
- **read_buffer_size**: Sequential scan buffer

### I/O Settings
- **innodb_flush_method**: I/O method optimization
- **innodb_log_file_size**: Transaction log sizing
- **innodb_flush_log_at_trx_commit**: Durability vs performance
- **sync_binlog**: Binary log synchronization

### Connection Settings
- **max_connections**: Maximum concurrent connections
- **max_user_connections**: Per-user connection limits
- **connect_timeout**: Connection establishment timeout
- **wait_timeout**: Idle connection timeout

## Troubleshooting

### Common Issues
1. **Permission denied**: Check file/directory permissions
2. **Cannot bind**: Port already in use or permission issues
3. **Memory issues**: Check buffer pool size and system memory
4. **Slow performance**: Review slow query log and configuration

### Debugging
- **Error logs**: Check `/var/log/mysql/error.log`
- **systemd status**: `systemctl status mariadb`
- **MySQL CLI**: `mariadb -u root -p` for connectivity test
- **Process list**: `SHOW PROCESSLIST` for active connections

### Monitoring Commands
```bash
# Check MariaDB status
mariadb -u root -p -e "SHOW STATUS"

# Get server variables
mariadb -u root -p -e "SHOW VARIABLES"

# Monitor slow queries
mariadb -u root -p -e "SHOW FULL PROCESSLIST"

# Check InnoDB status
mariadb -u root -p -e "SHOW ENGINE INNODB STATUS"

# Monitor connections
mariadb -u root -p -e "SHOW STATUS LIKE 'Threads%'"
```

## Version Compatibility

This configuration is optimized for:
- **MariaDB 11.2.2**: Primary supported version
- **MySQL Compatibility**: Most settings are MySQL compatible
- **Ubuntu 20.04+**: Tested on modern Ubuntu versions
- **Debian 10+**: Debian support included

## Migration and Upgrade

### From MySQL
- Configuration is largely MySQL compatible
- Binary log format compatible
- User accounts transferable with minor adjustments

### Version Upgrades
- Configuration validated for MariaDB 11.x series
- Backward compatible with MariaDB 10.x settings
- Forward compatible design for future versions

## Backup Integration

### Logical Backups
- **mysqldump**: Full database dumps
- **Character Set**: UTF-8 support for international data
- **Consistent Snapshots**: InnoDB consistent backup support

### Physical Backups
- **Binary Logs**: Point-in-time recovery
- **InnoDB Files**: Hot backup support
- **File System**: Snapshot-friendly configuration

### Replication Setup
- **Binary Logging**: Enabled for replication
- **Server ID**: Configurable for multi-master setups
- **GTID**: Global Transaction ID support
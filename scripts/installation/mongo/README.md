# MongoDB Configuration Files

This directory contains configuration templates for MongoDB installation.

## Files Overview

### Configuration Files
- `mongod.conf` - Host/native installation comprehensive configuration template
- `mongod-docker.conf` - Docker container installation optimized configuration template

### Service Files
- `mongod.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_MONGO_HOST` - MongoDB server host (default: 127.0.0.1)
- `PROJ_MONGO_PORT` - MongoDB server port (default: 27017)

### Authentication
- `PROJ_MONGO_ADMIN_USERNAME` - Admin username (default: root)
- `PROJ_MONGO_ADMIN_PASSWORD` - Admin password
- `PROJ_MONGO_DATABASE` - Default database name (default: proj)

### Directories
- `PROJ_MONGO_DATA_DIR` - Data directory (default: /var/lib/mongodb)
- `PROJ_MONGO_CONFIG_DIR` - Configuration directory (default: /etc)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. MongoDB runs as `mongodb` user with proper permissions
5. Admin user is created with specified credentials

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration is mounted into the container at `/etc/mongod.conf`
3. Data persisted to host directory via volume mount
4. Runs with optimized settings for containerized environment

## Configuration Types

### Host Installation (mongod.conf)
- **Production Ready**: Full MongoDB server configuration
- **System Integration**: Proper daemon mode, logging to files, PID management
- **Security**: Authentication enabled, bind to specific interface
- **Performance**: Optimized for production server environment
- **WiredTiger Engine**: Full WiredTiger storage engine configuration

### Docker Installation (mongod-docker.conf)
- **Container Optimized**: Adapted for containerized environment
- **Simplified Logging**: Console output for Docker log aggregation
- **Data Persistence**: Volume-mounted data directory (/data/db)
- **Network**: Binds to all interfaces (0.0.0.0) for container access
- **Resource Limits**: Conservative settings suitable for development

## Key Features Configured

### Storage Engine
- **WiredTiger**: High-performance storage engine
- **Compression**: Document and index compression
- **Cache Management**: Configurable cache size
- **Journaling**: Write-ahead logging for durability

### Security
- **Authentication**: SCRAM-SHA-1/256 authentication
- **Authorization**: Role-based access control
- **Network Binding**: Interface restrictions
- **SSL/TLS**: Optional encrypted connections

### Performance
- **Connection Pooling**: Efficient connection management
- **Index Management**: Automatic index optimization
- **Query Profiling**: Slow operation tracking
- **Oplog Sizing**: Replication log optimization

### Logging
- **System Log**: Comprehensive system logging
- **Access Log**: Connection and authentication logging
- **Slow Operations**: Performance monitoring
- **Log Rotation**: Automatic log file management

## Integration Points

### With Applications
- Standard MongoDB protocol on configured port
- Connection string: `mongodb://username:password@host:port/database`
- Multiple database support with proper isolation

### With Monitoring Stack
- **Prometheus**: MongoDB metrics can be exported via mongodb_exporter
- **Logging**: Structured logs for aggregation
- **Health Checks**: Built-in status commands for monitoring

### With Backup Systems
- **mongodump**: Logical backup utility
- **Oplog**: Point-in-time recovery support
- **File System**: Snapshot-friendly data files

## Directory Structure

After installation:

### Host Installation
```
/var/lib/mongodb/            # Data directory
├── collection-*.wt         # Collection data files
├── index-*.wt              # Index files
├── WiredTiger*             # WiredTiger metadata
├── mongod.lock             # Lock file
└── journal/                # Journal files

/var/log/mongodb/            # Log directory
├── mongod.log              # Main log file
└── ...

/etc/                        # Configuration directory
├── mongod.conf             # Main configuration
└── ...
```

### Docker Installation
```
/data/db/                    # Container data directory
├── collection-*.wt         # Collection data files
├── index-*.wt              # Index files
├── WiredTiger*             # WiredTiger metadata
└── journal/                # Journal files

/etc/mongod.conf            # Container configuration
```

## Security Considerations

### systemd Service Security
- **User Isolation**: Runs as dedicated `mongodb` user
- **File System Security**: Restricted access to data directories
- **Resource Limits**: Memory and connection limits
- **Process Isolation**: Proper process containment

### Database Security
- **Authentication**: Strong authentication mechanisms
- **Authorization**: Role-based access control
- **Network Security**: Bind address restrictions
- **SSL/TLS**: Optional encrypted connections

### File System Security
- **Directory Permissions**: Secure data directory access
- **Log Security**: Proper log file permissions
- **Configuration Protection**: Secure config file access

## Performance Tuning

### WiredTiger Settings
- **Cache Size**: Configure WiredTiger cache
- **Compression**: Document and index compression
- **Checkpoints**: Checkpoint interval tuning
- **Journal**: Journal commit interval

### Connection Settings
- **Max Connections**: Maximum concurrent connections
- **Connection Pool**: Connection pooling settings
- **Socket Timeout**: Connection timeout settings
- **Keep Alive**: TCP keep-alive settings

### Operation Settings
- **Slow Operation Threshold**: Slow query logging
- **Profiling**: Database profiling levels
- **Index Management**: Index build settings
- **Query Planning**: Query optimizer settings

## Troubleshooting

### Common Issues
1. **Permission denied**: Check file/directory permissions
2. **Cannot bind**: Port already in use or permission issues
3. **Authentication failed**: Check username/password
4. **Slow performance**: Review slow operation log

### Debugging
- **MongoDB logs**: Check `/var/log/mongodb/mongod.log`
- **systemd status**: `systemctl status mongod`
- **MongoDB shell**: `mongosh` for connectivity test
- **Current operations**: `db.currentOp()` for active operations

### Monitoring Commands
```bash
# Check MongoDB status
mongosh --eval "db.runCommand({serverStatus: 1})"

# Get database statistics
mongosh --eval "db.stats()"

# Monitor slow operations
mongosh --eval "db.getProfilingStatus()"

# Check replica set status (if applicable)
mongosh --eval "rs.status()"

# Monitor connections
mongosh --eval "db.serverStatus().connections"
```

## Version Compatibility

This configuration is optimized for:
- **MongoDB 7.0.x**: Primary supported version
- **WiredTiger**: Default and recommended storage engine
- **Ubuntu 20.04+**: Tested on modern Ubuntu versions
- **Debian 10+**: Debian support included

## Replication and Sharding

### Replica Set Configuration
- **Oplog Size**: Optimized for replication
- **Read Preference**: Configurable read preferences
- **Write Concern**: Configurable write durability
- **Election Timeout**: Replica set election settings

### Sharding Configuration
- **Config Servers**: Configuration server settings
- **Shard Servers**: Individual shard configuration
- **mongos Routers**: Query router configuration
- **Balancer**: Chunk migration settings

## Backup and Recovery

### Backup Strategies
- **mongodump**: Logical backup utility
- **File System Snapshots**: Point-in-time snapshots
- **Oplog Backup**: Incremental backup support
- **Cloud Backup**: Atlas backup integration

### Recovery Procedures
- **mongorestore**: Restore from logical backups
- **Oplog Replay**: Point-in-time recovery
- **Snapshot Restore**: File system restoration
- **Replica Recovery**: Automatic replica sync

## Migration and Upgrade

### Version Upgrades
- **Feature Compatibility**: Gradual feature adoption
- **Rolling Upgrades**: Zero-downtime upgrades
- **Index Rebuilds**: Automatic index optimization
- **Configuration Migration**: Setting compatibility

### Data Migration
- **mongoimport/mongoexport**: Data import/export
- **Change Streams**: Real-time data sync
- **Aggregation Pipelines**: Data transformation
- **Bulk Operations**: Efficient data loading
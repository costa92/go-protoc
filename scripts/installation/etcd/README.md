# etcd Configuration Files

This directory contains configuration templates for etcd installation.

## Files Overview

### Configuration Files
- `etcd.conf` - Host/native installation configuration template
- `etcd-docker.conf` - Docker container installation configuration template (for config file mounting)

### Service Files
- `etcd.service` - systemd service file template for native installation

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_ETCD_HOST` - etcd server host (default: 127.0.0.1)
- `PROJ_ETCD_PORT` - etcd client port (default: 2379)
- `PROJ_ETCD_PEER_PORT` - etcd peer port (default: 2380)

### Cluster Configuration
- `ETCD_NODE_NAME` - Node name (default: etcd-node1)
- `ETCD_CLUSTER_TOKEN` - Cluster token (default: etcd-cluster-1)
- `ETCD_CLUSTER_STATE` - Cluster state (default: new)

### Directories
- `PROJ_ETCD_DATA_DIR` - Data directory (default: /var/lib/etcd)

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Native Installation
1. Environment variables are substituted in templates
2. Processed config files are copied to system directories
3. systemd service is created and started
4. etcd runs as `etcd` user with proper permissions

### Docker Installation  
1. Environment variables are substituted in Docker-specific templates
2. Configuration can be mounted into the container (optional)
3. Data persisted to host directory via volume mount
4. Runs with optimized settings for containerized environment

## Configuration Types

### Host Installation (etcd.conf)
- **Production Ready**: Full etcd configuration for production use
- **System Integration**: Proper daemon mode, logging, PID management
- **Security**: Authentication and authorization options
- **Performance**: Optimized for production server environment
- **Clustering**: Multi-node cluster support

### Docker Installation (etcd-docker.conf)
- **Container Optimized**: Adapted for containerized environment
- **Simplified Logging**: Console output for Docker log aggregation
- **Data Persistence**: Volume-mounted data directory
- **Network**: Binds to all interfaces for container access
- **Single Node**: Optimized for development single-node setup

## Key Features Configured

### Core Features
- **Distributed Storage**: Reliable key-value storage
- **RAFT Consensus**: Strong consistency guarantee
- **Watch API**: Real-time change notifications
- **Transaction Support**: Multi-key atomic operations

### Clustering
- **Member Management**: Dynamic cluster membership
- **Leader Election**: Automatic leader selection
- **Data Replication**: Automatic data synchronization
- **Split-brain Prevention**: Quorum-based decisions

### Security
- **Client Authentication**: Optional TLS client certificates
- **Peer Authentication**: TLS for inter-node communication
- **Role-based Access**: Fine-grained permission control
- **Network Security**: Interface binding restrictions

### Performance
- **Write Batching**: Efficient write operations
- **Compaction**: Automatic history cleanup
- **Snapshot Management**: Periodic state snapshots
- **Memory Management**: Configurable memory limits

## Integration Points

### With Applications
- **HTTP API**: RESTful API on client port
- **gRPC API**: High-performance gRPC interface
- **etcdctl**: Command-line client tool
- **Client Libraries**: Native libraries for major languages

### With Monitoring Stack
- **Prometheus**: Built-in metrics endpoint
- **Health Checks**: Endpoint health monitoring
- **Member Status**: Cluster member monitoring
- **Performance Metrics**: Latency and throughput metrics

### With Service Discovery
- **Kubernetes**: Native etcd integration
- **Service Registration**: Dynamic service discovery
- **Configuration Management**: Centralized configuration
- **Leader Election**: Distributed coordination

## Directory Structure

After installation:

### Host Installation
```
/var/lib/etcd/               # Data directory
├── member/                  # Member data
│   ├── snap/               # Snapshots
│   └── wal/                # Write-ahead logs
└── ...

/etc/etcd/                   # Configuration directory (optional)
├── etcd.conf               # Configuration file
└── ...

/var/log/etcd/              # Log directory (if file logging enabled)
├── etcd.log                # Application logs
└── ...
```

### Docker Installation
```
/data/etcd/                  # Container data directory
├── member/                  # Member data
│   ├── snap/               # Snapshots
│   └── wal/                # Write-ahead logs
└── ...
```

## Security Considerations

### systemd Service Security
- **User Isolation**: Runs as dedicated `etcd` user
- **File System Security**: Restricted access to data directories
- **Resource Limits**: File descriptor and memory limits
- **Process Isolation**: Proper process containment

### Network Security
- **TLS Encryption**: Optional client and peer TLS
- **Authentication**: Client certificate authentication
- **Authorization**: Role-based access control
- **Interface Binding**: Restrict to specific interfaces

### Data Security
- **Encryption at Rest**: Optional data encryption
- **Secure Transport**: TLS for all communications
- **Access Control**: Fine-grained permissions
- **Audit Logging**: Security event logging

## Performance Tuning

### Memory Settings
- **Cache Size**: Configure member cache
- **Snapshot Count**: Snapshot frequency tuning
- **Compaction Mode**: Automatic vs manual compaction
- **Memory Limits**: Process memory constraints

### Disk I/O Settings
- **WAL Directory**: Separate WAL from data directory
- **Sync Settings**: Write synchronization modes
- **Disk Bandwidth**: I/O priority settings
- **Storage Backend**: Backend storage optimization

### Network Settings
- **Heartbeat Interval**: Member heartbeat frequency
- **Election Timeout**: Leader election timeout
- **Client Timeout**: Client connection timeout
- **Peer URLs**: Peer communication endpoints

## Troubleshooting

### Common Issues
1. **Split-brain**: Multiple leaders in cluster
2. **Network partitions**: Communication failures
3. **Disk space**: Insufficient storage space
4. **Performance**: High latency operations

### Debugging
- **etcd logs**: Check application logs
- **systemd status**: `systemctl status etcd`
- **Member health**: `etcdctl endpoint health`
- **Cluster status**: `etcdctl endpoint status`

### Monitoring Commands
```bash
# Check etcd version
etcdctl version

# Check cluster health
etcdctl --endpoints=http://127.0.0.1:2379 endpoint health

# Get cluster member list
etcdctl --endpoints=http://127.0.0.1:2379 member list

# Check endpoint status
etcdctl --endpoints=http://127.0.0.1:2379 endpoint status

# Monitor metrics
curl http://127.0.0.1:2379/metrics

# Test basic operations
etcdctl --endpoints=http://127.0.0.1:2379 put /test "hello"
etcdctl --endpoints=http://127.0.0.1:2379 get /test
```

## Version Compatibility

This configuration is optimized for:
- **etcd 3.5.x**: Primary supported version
- **API Version 3**: Latest etcd API
- **Ubuntu 20.04+**: Tested on modern Ubuntu versions
- **Debian 10+**: Debian support included

## Clustering Configuration

### Single Node (Development)
- **Cluster Size**: 1 node
- **Cluster State**: new
- **Initial Cluster**: single member
- **Data Safety**: No redundancy

### Three Node Cluster (Production)
- **Cluster Size**: 3 nodes (recommended minimum)
- **Cluster State**: new for bootstrap, existing for joins
- **Initial Cluster**: all three members
- **Data Safety**: Tolerates 1 node failure

### Five Node Cluster (High Availability)
- **Cluster Size**: 5 nodes (maximum recommended)
- **Cluster State**: new for bootstrap, existing for joins
- **Initial Cluster**: all five members
- **Data Safety**: Tolerates 2 node failures

## Backup and Recovery

### Backup Strategies
- **Snapshot Backup**: Point-in-time cluster snapshots
- **WAL Backup**: Write-ahead log backups
- **Automated Backup**: Periodic backup scheduling
- **Cross-region Backup**: Geographic distribution

### Recovery Procedures
- **Snapshot Restore**: Restore from cluster snapshot
- **Member Recovery**: Replace failed cluster member
- **Disaster Recovery**: Complete cluster rebuild
- **Data Verification**: Post-recovery validation

## Migration and Upgrade

### Version Upgrades
- **Rolling Upgrades**: Zero-downtime upgrades
- **Member by Member**: Sequential member updates
- **Compatibility Check**: Version compatibility verification
- **Rollback Plan**: Upgrade failure recovery

### Data Migration
- **etcdctl**: Data export/import utilities
- **Cluster Migration**: Move to new cluster
- **Backup Restore**: Migrate via backup/restore
- **Live Migration**: Online data migration
# Sentry Configuration Files

This directory contains configuration templates for Sentry installation.

## Files Overview

### Configuration Files
- `sentry.conf.py` - Main Sentry configuration template
- `config.yml` - Sentry YAML configuration template
- `docker-compose.yml` - Docker Compose configuration for complete Sentry stack

### Environment Files
- `.env` - Environment variables for Docker Compose setup

## Environment Variables

The following environment variables are used in templates and will be substituted by `envsubst`:

### Server Configuration
- `PROJ_SENTRY_HOST` - Sentry server host (default: 127.0.0.1)
- `PROJ_SENTRY_PORT` - Sentry server port (default: 9000)
- `PROJ_SENTRY_WEB_PORT` - Sentry web UI port (default: 9000)

### Database Configuration
- `PROJ_SENTRY_POSTGRES_DB` - PostgreSQL database name (default: sentry)
- `PROJ_SENTRY_POSTGRES_USER` - PostgreSQL user (default: sentry)
- `PROJ_SENTRY_POSTGRES_PASSWORD` - PostgreSQL password

### Cache Configuration
- `PROJ_SENTRY_REDIS_URL` - Redis URL for caching and queues
- `PROJ_SENTRY_REDIS_HOST` - Redis host
- `PROJ_SENTRY_REDIS_PORT` - Redis port (default: 6379)

### Security
- `PROJ_SENTRY_SECRET_KEY` - Sentry secret key for cryptographic operations

## Usage

These templates are automatically processed by the installation script using `envsubst` to substitute environment variables with their actual values.

### Docker Installation (Recommended)
1. Environment variables are substituted in Docker Compose templates
2. Complete stack deployed with dependencies (PostgreSQL, Redis)
3. Data persisted to host directories via volume mounts
4. Production-ready configuration with proper networking

### Native Installation (Development Only)
1. Environment variables are substituted in Python configuration
2. Manual dependency management required
3. Simplified setup for development and testing
4. Not recommended for production use

## Configuration Types

### Docker Compose Setup (docker-compose.yml)
- **Complete Stack**: Sentry, PostgreSQL, Redis, Kafka (optional)
- **Production Ready**: Proper resource limits and health checks
- **Data Persistence**: Volume mounts for all stateful services
- **Network Isolation**: Dedicated Docker network
- **Load Balancing**: Multiple Sentry workers for scalability

### Python Configuration (sentry.conf.py)
- **Database Settings**: PostgreSQL connection configuration
- **Cache Configuration**: Redis cache and session storage
- **Email Settings**: SMTP configuration for notifications
- **Security Settings**: Secret keys and authentication
- **Feature Flags**: Optional Sentry features

### YAML Configuration (config.yml)
- **Service Discovery**: Kubernetes-ready configuration
- **Environment Separation**: Dev/staging/production configs
- **Resource Limits**: Memory and CPU constraints
- **Scaling Options**: Horizontal scaling configuration

## Key Features Configured

### Error Tracking
- **Exception Capture**: Automatic error collection
- **Source Maps**: JavaScript source map support
- **Release Tracking**: Version-based error grouping
- **Issue Management**: Automated issue creation and assignment

### Performance Monitoring
- **Transaction Tracing**: End-to-end request tracking
- **Performance Metrics**: Latency and throughput monitoring
- **Custom Metrics**: Application-specific measurements
- **Alert Rules**: Performance-based alerting

### Security
- **Authentication**: SAML, OAuth, LDAP integration
- **Authorization**: Role-based access control
- **Data Scrubbing**: Sensitive data filtering
- **Rate Limiting**: API and UI rate limiting

### Integrations
- **Version Control**: GitHub, GitLab, Bitbucket
- **Issue Tracking**: Jira, GitHub Issues, Linear
- **Communication**: Slack, Microsoft Teams, PagerDuty
- **CI/CD**: Jenkins, GitHub Actions, GitLab CI

## Integration Points

### With Applications
- **SDK Integration**: Multi-language SDK support
- **Error Capture**: Automatic exception reporting
- **Performance Monitoring**: Request performance tracking
- **Release Management**: Deploy and release tracking

### With Monitoring Stack
- **Prometheus**: Custom metrics export
- **Grafana**: Performance dashboards
- **Health Checks**: Service health monitoring
- **Log Aggregation**: Centralized log collection

### With CI/CD Pipelines
- **Release Creation**: Automated release tracking
- **Deploy Notifications**: Deployment event capture
- **Source Map Upload**: Automated source map deployment
- **Issue Assignment**: Automatic bug assignment

## Directory Structure

After installation:

### Docker Installation
```
/data/sentry/                    # Sentry data directory
├── postgres/                    # PostgreSQL data
├── redis/                       # Redis data
├── sentry/                      # Sentry application data
│   ├── files/                   # Uploaded files
│   └── data/                    # Application data
└── config/                      # Configuration files
    ├── sentry.conf.py          # Main configuration
    ├── config.yml              # YAML configuration
    └── .env                    # Environment variables
```

### Native Installation
```
/etc/sentry/                     # Configuration directory
├── sentry.conf.py              # Main configuration
├── config.yml                 # YAML configuration
└── .env                       # Environment variables

/var/lib/sentry/                # Data directory
├── files/                      # Uploaded files
└── data/                      # Application data

/var/log/sentry/               # Log directory
├── sentry.log                 # Application logs
└── worker.log                 # Background worker logs
```

## Security Considerations

### Application Security
- **Secret Management**: Secure secret key generation
- **Database Security**: Encrypted database connections
- **Session Security**: Secure session management
- **Data Privacy**: PII scrubbing and anonymization

### Network Security
- **TLS Encryption**: HTTPS for web interface
- **API Security**: Rate limiting and authentication
- **Container Security**: Minimal container images
- **Network Isolation**: Dedicated container networks

### Data Security
- **Backup Encryption**: Encrypted backup storage
- **Access Control**: Role-based permissions
- **Audit Logging**: Security event tracking
- **Compliance**: GDPR and SOC 2 compliance features

## Performance Tuning

### Application Settings
- **Worker Configuration**: Background worker tuning
- **Cache Settings**: Redis cache optimization
- **Database Settings**: PostgreSQL performance tuning
- **Queue Management**: Celery queue configuration

### Resource Allocation
- **Memory Limits**: Container memory constraints
- **CPU Limits**: Processing power allocation
- **Storage**: Database and file storage sizing
- **Network**: Bandwidth and connection limits

### Scaling Configuration
- **Horizontal Scaling**: Multiple application instances
- **Load Balancing**: Request distribution
- **Database Scaling**: Read replicas and sharding
- **Cache Scaling**: Redis clustering

## Troubleshooting

### Common Issues
1. **Database Connection**: PostgreSQL connectivity issues
2. **Redis Connection**: Cache and queue connectivity
3. **Memory Issues**: Insufficient memory allocation
4. **Performance**: High latency or timeouts

### Debugging
- **Application Logs**: Sentry application logging
- **Worker Logs**: Background task debugging
- **Database Logs**: PostgreSQL query logging
- **Container Logs**: Docker container diagnostics

### Monitoring Commands
```bash
# Check Sentry web interface
curl -s http://127.0.0.1:9000/_health/

# Check database connectivity
docker exec sentry-postgres psql -U sentry -d sentry -c "SELECT 1"

# Check Redis connectivity
docker exec sentry-redis redis-cli ping

# View Sentry logs
docker logs sentry-web

# Check worker status
docker logs sentry-worker

# Monitor resource usage
docker stats sentry-web sentry-worker sentry-postgres sentry-redis
```

## Version Compatibility

This configuration is optimized for:
- **Sentry 24.x**: Latest stable version
- **PostgreSQL 15**: Recommended database version
- **Redis 7**: Latest cache and queue backend
- **Python 3.11**: Runtime environment

## Deployment Strategies

### Development Setup
- **Single Container**: Simplified setup for development
- **SQLite Database**: File-based database for testing
- **Minimal Resources**: Low resource requirements
- **Quick Setup**: Fast installation and configuration

### Production Setup
- **Multi-Container**: Separate services for scalability
- **PostgreSQL**: Production-grade database
- **Redis Cluster**: High-availability caching
- **Load Balancer**: Request distribution and failover

### High Availability
- **Database Replication**: PostgreSQL master-slave setup
- **Redis Sentinel**: Redis high availability
- **Multiple Workers**: Distributed background processing
- **Health Monitoring**: Automated failure detection

## Backup and Recovery

### Backup Strategies
- **Database Backup**: PostgreSQL dump and restore
- **File Backup**: Uploaded files and attachments
- **Configuration Backup**: Settings and customizations
- **Automated Backup**: Scheduled backup jobs

### Recovery Procedures
- **Database Recovery**: Point-in-time recovery
- **File Recovery**: File system restoration
- **Configuration Recovery**: Settings restoration
- **Disaster Recovery**: Complete system rebuild

## Migration and Upgrade

### Version Upgrades
- **Database Migration**: Automatic schema updates
- **Configuration Migration**: Settings compatibility
- **Data Migration**: User data preservation
- **Rollback Capability**: Upgrade failure recovery

### Data Migration
- **Import/Export**: Data transfer utilities
- **Cross-Instance**: Migration between Sentry instances
- **Bulk Operations**: Large dataset migration
- **Validation**: Post-migration data verification
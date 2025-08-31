# Redis configuration template for Ubuntu native deployment
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

# Network Configuration
bind 127.0.0.1 ::1
port ${PROJ_REDIS_PORT:-6379}
protected-mode yes

# General Configuration
daemonize yes
pidfile /run/redis/redis-server.pid
loglevel notice
logfile ${PROJ_REDIS_LOG_DIR:-/var/log/redis}/redis-server.log

# Persistence Configuration
save 900 1
save 300 10
save 60 10000
rdbcompression yes
dbfilename dump.rdb
dir ${PROJ_REDIS_DATA_DIR:-/var/lib/redis}

# Memory Configuration
maxmemory ${REDIS_MAX_MEMORY:-256mb}
maxmemory-policy allkeys-lru

# Security Configuration
# requirepass ${REDIS_PASSWORD:-}

# Slow Log Configuration
slowlog-log-slower-than 10000
slowlog-max-len 128

# Client Configuration
timeout 300
tcp-keepalive 300
tcp-backlog 511

# Ubuntu specific settings
supervised systemd
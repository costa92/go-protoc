# Redis configuration template for Docker deployment
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

# Network Configuration
bind 0.0.0.0
port 6379
protected-mode no

# General Configuration
daemonize no
pidfile /var/run/redis.pid
loglevel notice
logfile ""

# Persistence Configuration
save 900 1
save 300 10
save 60 10000
rdbcompression yes
dbfilename dump.rdb
dir /data

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
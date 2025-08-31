# Redis configuration template for macOS native deployment
# Project: ${PROJ_NAME:-go-protoc}
# Service: Redis ${REDIS_VERSION}
# Environment: ${PROJ_ENVIRONMENT:-development}

# Network Configuration
bind 127.0.0.1
port ${PROJ_REDIS_PORT:-6379}
protected-mode yes

# General Configuration
daemonize no
pidfile ${PROJ_REDIS_DATA_DIR}/redis.pid
loglevel notice
logfile ${PROJ_REDIS_LOG_DIR:-${PROJ_REDIS_DATA_DIR}}/redis.log

# Persistence Configuration
save 900 1
save 300 10
save 60 10000
rdbcompression yes
dbfilename dump.rdb
dir ${PROJ_REDIS_DATA_DIR}

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

# macOS specific settings
unixsocket ${PROJ_REDIS_DATA_DIR}/redis.sock
unixsocketperm 700
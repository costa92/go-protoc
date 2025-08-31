[Unit]
Description=Redis In-Memory Data Store (${REDIS_VERSION})
Documentation=https://redis.io/documentation
After=network.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/redis-server ${PROJ_REDIS_CONFIG_DIR}/redis.conf
ExecReload=/bin/kill -HUP $MAINPID
KillMode=mixed
Restart=always
RestartSec=5
TimeoutStopSec=20
User=redis
Group=redis
RuntimeDirectory=redis
RuntimeDirectoryMode=0755

# Security Settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=${PROJ_REDIS_DATA_DIR} ${PROJ_REDIS_CONFIG_DIR}

# Resource Limits
LimitNOFILE=65535
LimitNPROC=32768

# Working Directory
WorkingDirectory=${PROJ_REDIS_DATA_DIR}

# Environment Variables
Environment=HOME=${PROJ_REDIS_DATA_DIR}
Environment=PROJ_SERVICE_NAME=redis
Environment=PROJ_SERVICE_VERSION=${REDIS_VERSION}

[Install]
WantedBy=multi-user.target
#!/bin/bash

# 等待MySQL就绪的脚本
set -e

HOST=${1:-127.0.0.1}
PORT=${2:-3306}
USER=${3:-root}
PASSWORD=${4:-proj\(\#\)666}
DATABASE=${5:-onex}
TIMEOUT=${6:-60}

echo "Waiting for MySQL to be ready at $HOST:$PORT..."
echo "Database: $DATABASE"
echo "Timeout: ${TIMEOUT}s"

counter=0
while ! mysql -h"$HOST" -P"$PORT" -u"$USER" -p"$PASSWORD" -e "SELECT 1" "$DATABASE" >/dev/null 2>&1; do
    if [ $counter -ge $TIMEOUT ]; then
        echo "❌ MySQL failed to start within ${TIMEOUT} seconds"
        exit 1
    fi
    
    echo "⏳ Waiting for MySQL... (${counter}/${TIMEOUT}s)"
    sleep 1
    counter=$((counter + 1))
done

echo "✅ MySQL is ready!"
echo "📊 MySQL version:"
mysql -h"$HOST" -P"$PORT" -u"$USER" -p"$PASSWORD" -e "SELECT VERSION();" "$DATABASE"
echo ""
echo "📋 Database status:"
mysql -h"$HOST" -P"$PORT" -u"$USER" -p"$PASSWORD" -e "SHOW DATABASES;" 2>/dev/null | grep -E "^(onex|mysql|information_schema|performance_schema|sys)$"
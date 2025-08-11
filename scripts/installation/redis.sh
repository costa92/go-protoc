#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


# Set some environment variables.
PROJ_REDIS_HOST=${PROJ_REDIS_HOST:-127.0.0.1}
PROJ_REDIS_PORT=${PROJ_REDIS_PORT:-6379}
PROJ_REDIS_PASSWORD=${PROJ_REDIS_PASSWORD:-proj(#)666}


# Function to install Redis using kubectl
proj::redis::install() {
    proj::log::info "Installing Redis..."

        # Check if kubectl is available
    if ! proj::util::cmd_exists kubectl; then
        proj::log::error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi

    # Apply Redis deployment and service
    kubectl apply -f ${PROJ_ROOT_DIR}/deployments/redis/redis.yaml

    # Wait for Redis pod to be ready
    proj::log::info "Waiting for Redis pod to be ready..."
    proj::kubectl wait --for=condition=ready pod -l app=redis --timeout=120s

    proj::log::info "Redis installation completed successfully!"
}

# Pre-install Redis by system init and package management.
proj::redis::pre_install(){
    proj::log::info "Pre-installing Redis..."
    # 判断是 mac 还是 linux
    if proj::util::is_mac; then
        proj::log::info "Mac OS detected, skipping Redis installation..."
        # proj::util::exec "brew install redis-cli"
    else
        proj::util::sudo "apt install -y redis-tools"
    fi
}

# Func
proj::redis::docker::install(){
    proj::log::info "Installing docker Redis..."

    proj::redis::pre_install
    proj::common::network

    docker run -d --name proj-redis \
      --restart always \
      --network ${NETWORK_NAME} \
      -v ${PROJ_THIRDPARTY_INSTALL_DIR}/redis:/data \
      -p ${PROJ_REDIS_HOST}:${PROJ_REDIS_PORT}:6379 \
      redis:7.2.3 \
      redis-server \
      --appendonly yes \
      --save 60 1 \
      --protected-mode no \
      --requirepass ${PROJ_REDIS_PASSWORD} \
      --loglevel debug

    sleep 2
    if proj::util::is_linux; then
        proj::redis::status || return 1
    fi
    proj::redis::info
    proj::log::info "install redis successfully"
}


# Print necessary information after docker or sbs installation.
proj::redis::info()
{
  echo -e ${C_GREEN}redis has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Redis access endpoint is: ${PROJ_REDIS_HOST}:${PROJ_REDIS_PORT}
       Redis password is: ${PROJ_REDIS_PASSWORD}
     Redis Login Command: redis-cli --no-auth-warning -h ${PROJ_REDIS_HOST} -p ${PROJ_REDIS_PORT} -a '${PROJ_REDIS_PASSWORD}'
EOF
}

# Uninstall the docker container.
proj::redis::docker::uninstall()
{
  docker rm -f proj-redis &>/dev/null
  proj::util::sudo "rm -rf ${PROJ_THIRDPARTY_INSTALL_DIR}/redis"
  proj::log::info "uninstall redis successfully"
}

# Status check after docker or sbs installation.
proj::redis::status()
{
  proj::util::telnet ${PROJ_REDIS_HOST} ${PROJ_REDIS_PORT} || return 1
  redis-cli --no-auth-warning -h ${PROJ_REDIS_HOST} -p ${PROJ_REDIS_PORT} -a "${PROJ_REDIS_PASSWORD}" --hotkeys || {
    proj::log::error "can not login with ${PROJ_REDIS_USERNAME}, redis maybe not initialized properly."
    return 1
  }
}

if [[ "$*" =~ proj::redis:: ]]; then
  eval $*
fi





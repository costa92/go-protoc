#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


# Function to install Redis using kubectl
proj::redis::install() {
    log::info "Installing Redis..."

        # Check if kubectl is available
    if ! util::cmd_exists kubectl; then
        log::error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi

    # Apply Redis deployment and service
    kubectl apply -f ${PROJ_ROOT_DIR}/deployments/redis/redis.yaml

    # Wait for Redis pod to be ready
    log::info "Waiting for Redis pod to be ready..."
    kubectl wait --for=condition=ready pod -l app=redis --timeout=120s

    log::info "Redis installation completed successfully!"
}

if [[ "$*" =~ proj::redis:: ]]; then
  eval $*
fi






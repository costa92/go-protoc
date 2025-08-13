#!/usr/bin/env bash

# This script installs the dependencies for the project.

set -o errexit
set -o nounset
set -o pipefail

# The root of the build/dist directory
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED:-} ]] && source ${PROJ_ROOT_DIR}/scripts/installation/common.sh
# Set some environment variables.
INSTALL_DIR=${PROJ_ROOT_DIR}/scripts/installation

source ${INSTALL_DIR}/redis.sh
source ${INSTALL_DIR}/mariadb.sh
source ${INSTALL_DIR}/mongo.sh
source ${INSTALL_DIR}/kafka.sh
source ${INSTALL_DIR}/etcd.sh
source ${INSTALL_DIR}/jaeger.sh


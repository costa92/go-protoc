#!/usr/bin/env bash

# Common utilities, variables and checks for all build scripts.
set -eEuo pipefail

# Unset CDPATH, having it set messes up with script import paths
unset CDPATH

# Colors
export BLUE='\033[0;34m'
export GREEN='\033[0;32m'
export RED='\033[0;31m'
export YELLOW='\033[0;33m'
export NC='\033[0m' # No Color


PROJ_VERBOSE=${PROJ_VERBOSE:-1}

# This will canonicalize the path
PROJ_ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd -P)
source "${PROJ_ROOT_DIR}/scripts/lib/init.sh"


# Set a flag to indicate this file has been sourced
export COMMON_SOURCED=true
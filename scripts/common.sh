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

# Project root directory
export PROJ_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"

# Function to print colored messages
function log::info() {
    echo -e "${GREEN}[INFO]${NC} $*"
}

function log::warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

function log::error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

function log::debug() {
    echo -e "${BLUE}[DEBUG]${NC} $*"
}

# Check if a command exists
function util::cmd_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if we're running on macOS
function util::is_darwin() {
    [[ "$(uname)" == "Darwin" ]]
}

# Check if we're running on Linux
function util::is_linux() {
    [[ "$(uname)" == "Linux" ]]
}

# Set a flag to indicate this file has been sourced
export COMMON_SOURCED=true
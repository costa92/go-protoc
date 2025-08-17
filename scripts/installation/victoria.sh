#!/usr/bin/env bash
#
# Victoria Suite Installation Script (Unified Entry Point)
# This script provides a unified entry point for managing VictoriaMetrics suite components:
# - VictoriaMetrics (time series database)
# - VictoriaLogs (log database)  
# - vmagent (metrics collection agent)
# - Victoria Suite (complete stack)
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

# 加载套件管理脚本
source "${SCRIPT_DIR}/victoria/victoria-suite.sh"

# =============================================================================
# Victoria Suite Management Functions (Complete Stack)
# =============================================================================

# Function to install complete Victoria suite
proj::victoria::install() {
  proj::victoria::suite::install
}

# Function to uninstall complete Victoria suite
proj::victoria::uninstall() {
  proj::victoria::suite::uninstall
}

# Function to install complete Victoria suite using Docker
proj::victoria::docker::install() {
  proj::victoria::suite::docker::install
}

# Function to uninstall complete Victoria suite Docker containers
proj::victoria::docker::uninstall() {
  proj::victoria::suite::docker::uninstall
}

# Function to check status of complete Victoria suite
proj::victoria::status() {
  proj::victoria::suite::status
}

# Function to display Victoria suite information
proj::victoria::info() {
  proj::victoria::suite::info
}

# Function to install all individual components using Docker
proj::victoria::install::all() {
  proj::log::info "Installing all Victoria components individually using Docker..."
  proj::log::info "================================================================"
  
  # Install each component individually
  proj::victoriametrics::docker::install
  echo ""
  proj::victorialogs::docker::install
  echo ""
  proj::vmagent::docker::install
  
  proj::log::info "All Victoria components installed successfully!"
  proj::victoria::info
}

# Function to uninstall all individual components
proj::victoria::uninstall::all() {
  proj::log::info "Uninstalling all Victoria components..."
  proj::log::info "======================================"
  
  # Uninstall in reverse order
  proj::vmagent::docker::uninstall
  echo ""
  proj::victorialogs::docker::uninstall
  echo ""
  proj::victoriametrics::docker::uninstall
  
  proj::log::info "All Victoria components uninstalled successfully!"
}

# =============================================================================
# Individual Component Management Functions (Legacy Compatibility)
# =============================================================================

# VictoriaMetrics individual component functions
proj::victoriametrics::install() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" install
}

proj::victoriametrics::uninstall() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" uninstall
}

proj::victoriametrics::docker::install() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" docker.install
}

proj::victoriametrics::docker::uninstall() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" docker.uninstall
}

proj::victoriametrics::status() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" status
}

proj::victoriametrics::info() {
  bash "${SCRIPT_DIR}/victoria/victoriametrics.sh" info
}

# VictoriaLogs individual component functions
proj::victorialogs::install() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" install
}

proj::victorialogs::uninstall() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" uninstall
}

proj::victorialogs::docker::install() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" docker.install
}

proj::victorialogs::docker::uninstall() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" docker.uninstall
}

proj::victorialogs::status() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" status
}

proj::victorialogs::info() {
  bash "${SCRIPT_DIR}/victoria/victorialogs.sh" info
}

# vmagent individual component functions
proj::vmagent::install() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" install
}

proj::vmagent::uninstall() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" uninstall
}

proj::vmagent::docker::install() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" docker.install
}

proj::vmagent::docker::uninstall() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" docker.uninstall
}

proj::vmagent::status() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" status
}

proj::vmagent::info() {
  bash "${SCRIPT_DIR}/victoria/vmagent.sh" info
}

# =============================================================================
# Command Line Interface
# =============================================================================

# Function to display usage information
proj::victoria::usage() {
  cat << EOF
Victoria Suite Installation Script

USAGE:
  $(basename "$0") COMMAND [OPTIONS]

COMMANDS:
  Suite Management (Complete Stack):
    install               Install complete Victoria suite natively
    uninstall            Uninstall complete Victoria suite
    docker.install       Install Victoria suite using Docker
    docker.uninstall     Uninstall Victoria suite Docker containers
    install.all          Install ALL individual components using Docker
    uninstall.all        Uninstall ALL individual components
    status               Check status of all components
    info                 Display suite information

  Individual Component Management:
    victoriametrics.COMMAND    Manage VictoriaMetrics component
    victorialogs.COMMAND       Manage VictoriaLogs component
    vmagent.COMMAND            Manage vmagent component

  Where COMMAND can be:
    install, uninstall, docker.install, docker.uninstall, install.all, uninstall.all, status, info

EXAMPLES:
  # Install complete suite using Docker
  $(basename "$0") docker.install

  # Install all individual components
  $(basename "$0") install.all

  # Uninstall all components
  $(basename "$0") uninstall.all

  # Check status of all components
  $(basename "$0") status

  # Install only VictoriaMetrics
  $(basename "$0") victoriametrics.docker.install

  # Check VictoriaLogs status
  $(basename "$0") victorialogs.status

  # Get suite information
  $(basename "$0") info

COMPONENTS:
  VictoriaMetrics    Time series database for metrics storage
  VictoriaLogs       Log database for log aggregation and search
  vmagent            Metrics collection and forwarding agent

MANAGEMENT COMMANDS:
  make deploy.install.docker.victoria          # Install complete suite
  make deploy.install.all.victoria             # Install all components individually
  make deploy.uninstall.all.victoria           # Uninstall all components
  make deploy.install.docker.victoriametrics   # Install VictoriaMetrics only
  make deploy.install.docker.victorialogs      # Install VictoriaLogs only
  make deploy.install.docker.vmagent           # Install vmagent only
  make deploy.status.victoria                  # Check suite status

EOF
}

# Main function to handle command line arguments
main() {
  local command=${1:-}
  
  case "${command}" in
    # Suite management commands
    install)
      proj::victoria::install
      ;;
    uninstall)
      proj::victoria::uninstall
      ;;
    docker.install)
      proj::victoria::docker::install
      ;;
    docker.uninstall)
      proj::victoria::docker::uninstall
      ;;
    status)
      proj::victoria::status
      ;;
    info)
      proj::victoria::info
      ;;
    install.all)
      proj::victoria::install::all
      ;;
    uninstall.all)
      proj::victoria::uninstall::all
      ;;
    
    # VictoriaMetrics individual component commands
    victoriametrics.install)
      proj::victoriametrics::install
      ;;
    victoriametrics.uninstall)
      proj::victoriametrics::uninstall
      ;;
    victoriametrics.docker.install)
      proj::victoriametrics::docker::install
      ;;
    victoriametrics.docker.uninstall)
      proj::victoriametrics::docker::uninstall
      ;;
    victoriametrics.status)
      proj::victoriametrics::status
      ;;
    victoriametrics.info)
      proj::victoriametrics::info
      ;;
    
    # VictoriaLogs individual component commands
    victorialogs.install)
      proj::victorialogs::install
      ;;
    victorialogs.uninstall)
      proj::victorialogs::uninstall
      ;;
    victorialogs.docker.install)
      proj::victorialogs::docker::install
      ;;
    victorialogs.docker.uninstall)
      proj::victorialogs::docker::uninstall
      ;;
    victorialogs.status)
      proj::victorialogs::status
      ;;
    victorialogs.info)
      proj::victorialogs::info
      ;;
    
    # vmagent individual component commands
    vmagent.install)
      proj::vmagent::install
      ;;
    vmagent.uninstall)
      proj::vmagent::uninstall
      ;;
    vmagent.docker.install)
      proj::vmagent::docker::install
      ;;
    vmagent.docker.uninstall)
      proj::vmagent::docker::uninstall
      ;;
    vmagent.status)
      proj::vmagent::status
      ;;
    vmagent.info)
      proj::vmagent::info
      ;;
    
    # Help and usage
    help|--help|-h)
      proj::victoria::usage
      ;;
    
    *)
      if [[ -z "${command}" ]]; then
        proj::log::error "No command specified"
      else
        proj::log::error "Unknown command: ${command}"
      fi
      echo ""
      proj::victoria::usage
      return 1
      ;;
  esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
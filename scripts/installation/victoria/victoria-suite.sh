#!/usr/bin/env bash
#
# VictoriaMetrics Suite Management Script
# This script provides unified management for VictoriaMetrics suite components:
# - VictoriaMetrics (time series database)
# - VictoriaLogs (log database)  
# - vmagent (metrics collection agent)
#

# Exit on any error, undefined variables, or pipe failures
set -o errexit
set -o nounset
set -o pipefail

# 加载通用配置和版本管理
SUITE_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SUITE_SCRIPT_DIR}/../common.sh"

# 加载各组件脚本
source "${SUITE_SCRIPT_DIR}/victoriametrics.sh"
source "${SUITE_SCRIPT_DIR}/victorialogs.sh"
source "${SUITE_SCRIPT_DIR}/vmagent.sh"

# Function to install complete VictoriaMetrics suite
proj::victoria::suite::install() {
  proj::log::info "Installing VictoriaMetrics Suite (native installation)..."
  
  # 安装各组件
  proj::victoriametrics::install
  proj::victorialogs::install
  proj::vmagent::install
  
  proj::log::info "VictoriaMetrics Suite installation completed"
  proj::victoria::suite::info
}

# Function to uninstall complete VictoriaMetrics suite
proj::victoria::suite::uninstall() {
  proj::log::info "Uninstalling VictoriaMetrics Suite..."
  
  # 卸载各组件
  proj::vmagent::uninstall
  proj::victorialogs::uninstall
  proj::victoriametrics::uninstall
  
  proj::log::info "VictoriaMetrics Suite uninstalled successfully"
}

# Function to install complete VictoriaMetrics suite using Docker
proj::victoria::suite::docker::install() {
  proj::log::info "Installing VictoriaMetrics Suite using Docker..."
  proj::common::network
  
  # 按依赖顺序安装
  proj::victoriametrics::docker::install
  proj::victorialogs::docker::install
  proj::vmagent::docker::install
  
  proj::log::info "VictoriaMetrics Suite Docker installation completed"
  proj::victoria::suite::info
}

# Function to uninstall complete VictoriaMetrics suite Docker containers
proj::victoria::suite::docker::uninstall() {
  proj::log::info "Uninstalling VictoriaMetrics Suite Docker containers..."
  
  # 按相反依赖顺序卸载
  proj::vmagent::docker::uninstall
  proj::victorialogs::docker::uninstall
  proj::victoriametrics::docker::uninstall
  
  proj::log::info "VictoriaMetrics Suite Docker containers removed"
}

# Function to check status of all components
proj::victoria::suite::status() {
  proj::log::info "Checking VictoriaMetrics Suite status..."
  proj::log::info "==============================================="
  
  proj::victoriametrics::status
  echo ""
  proj::victorialogs::status  
  echo ""
  proj::vmagent::status
}

# Function to display suite information
proj::victoria::suite::info() {
  proj::log::info "VictoriaMetrics Suite Information:"
  proj::log::info "=================================="
  proj::log::info ""
  proj::log::info "🎯 Complete Monitoring Stack:"
  proj::log::info "  VictoriaMetrics: Time series database and metrics storage"
  proj::log::info "  VictoriaLogs:    Log database and log aggregation"
  proj::log::info "  vmagent:         Metrics collection and forwarding agent"
  proj::log::info ""
  proj::log::info "🌐 Access URLs:"
  proj::log::info "  VictoriaMetrics UI:  http://127.0.0.1:8428"
  proj::log::info "  VictoriaLogs UI:     http://127.0.0.1:9428/select/vmui"
  proj::log::info "  vmagent UI:          http://127.0.0.1:8429"
  proj::log::info ""
  proj::log::info "📊 API Endpoints:"
  proj::log::info "  Metrics Query:       http://127.0.0.1:8428/api/v1/query"
  proj::log::info "  Metrics Insert:      http://127.0.0.1:8428/api/v1/write"
  proj::log::info "  Logs Query:          http://127.0.0.1:9428/select/logsql/query"
  proj::log::info "  Logs Insert:         http://127.0.0.1:9428/insert/jsonline"
  proj::log::info ""
  proj::log::info "🔧 Management Commands:"
  proj::log::info "  Suite status:        make deploy.status.victoria-suite"
  proj::log::info "  Component status:    make deploy.status.{victoriametrics|victorialogs|vmagent}"
  proj::log::info "  Restart all:         sudo systemctl restart victoriametrics victorialogs vmagent"
  proj::log::info ""
}

# Main function to handle command line arguments
main() {
  local action=${1:-}
  
  case "${action}" in
    install)
      proj::victoria::suite::install
      ;;
    uninstall)
      proj::victoria::suite::uninstall
      ;;
    docker.install)
      proj::victoria::suite::docker::install
      ;;
    docker.uninstall)
      proj::victoria::suite::docker::uninstall
      ;;
    status)
      proj::victoria::suite::status
      ;;
    info)
      proj::victoria::suite::info
      ;;
    *)
      proj::log::error "Usage: $0 {install|uninstall|docker.install|docker.uninstall|status|info}"
      proj::log::info ""
      proj::log::info "Individual component management:"
      proj::log::info "  ./victoriametrics.sh {install|uninstall|docker.install|docker.uninstall|status|info}"
      proj::log::info "  ./victorialogs.sh {install|uninstall|docker.install|docker.uninstall|status|info}"  
      proj::log::info "  ./vmagent.sh {install|uninstall|docker.install|docker.uninstall|status|info}"
      return 1
      ;;
  esac
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
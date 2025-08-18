#!/usr/bin/env bash

# This script manages the lifecycle of dependent services using docker-compose and installation scripts.

set -o errexit
set -o nounset
set -o pipefail

# Define the root directory of the project
# This allows the script to be run from anywhere
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/../.. && pwd)"

# Source versions and common functions
source "${ROOT_DIR}/scripts/installation/versions.sh"
source "${ROOT_DIR}/scripts/installation/common.sh"

# Function to print usage
usage() {
  echo "Usage: $0 <action> <service>"
  echo ""
  echo "Actions:"
  echo "  start      Start the specified service."
  echo "  stop       Stop and remove the specified service."
  echo "  status     Check service status."
  echo "  restart    Restart the specified service."
  echo "  logs       Show service logs."
  echo ""
  echo "Services:"
  echo "  Database:"
  echo "    redis        Redis cache service (docker)"
  echo "    mariadb      MariaDB database service (docker)"
  echo "    mongodb      MongoDB database service (docker)"
  echo "  Messaging:"
  echo "    kafka        Kafka messaging service (docker)"
  echo "  Distributed:"
  echo "    etcd         etcd distributed key-value store (binary)"
  echo "  Observability:"
  echo "    jaeger       Jaeger tracing service (docker)"
  echo "    prometheus   Prometheus monitoring (docker)"
  echo "    grafana      Grafana dashboard (docker)"
  echo "    alertmanager AlertManager for Prometheus (docker)"
  echo "    otelcol      OpenTelemetry Collector (docker)"
  echo "    victorialogs VictoriaLogs for log management (docker)"
  echo "  Groups:"
  echo "    all          Apply the action to all services."
  echo "    database     Apply to all database services (redis, mariadb, mongodb)"
  echo "    observability Apply to all observability services (jaeger, prometheus, grafana, etc.)"
  exit 1
}

# Check for the correct number of arguments
if [ "$#" -ne 2 ]; then
  usage
fi

ACTION=$1
SERVICE=$2

# Define service groups
DATABASE_SERVICES=("redis" "mariadb" "mongodb")
OBSERVABILITY_SERVICES=("jaeger" "prometheus" "grafana" "alertmanager" "otelcol" "victorialogs")
ALL_SERVICES=("${DATABASE_SERVICES[@]}" "kafka" "etcd" "${OBSERVABILITY_SERVICES[@]}")

# Service type mappings
declare -A SERVICE_TYPES=(
  ["redis"]="script"
  ["kafka"]="script" 
  ["jaeger"]="script"
  ["mariadb"]="script"
  ["mongodb"]="script"
  ["etcd"]="script"
  ["prometheus"]="script"
  ["grafana"]="script"
  ["alertmanager"]="script"
  ["otelcol"]="script"
  ["victorialogs"]="script"
)

# Function to manage services by docker-compose (deprecated - kept for reference)
# All services now use individual scripts for better control and consistency
manage_docker_compose_service() {
  echo "Error: docker-compose management is deprecated. All services now use individual scripts."
  echo "This function should not be called."
  return 1
}

# Function to manage services by installation scripts
manage_script_service() {
  local service_name=$1
  local action=$2
  # Handle special case where service name differs from script name
  local script_name=$service_name
  if [ "$service_name" = "mongodb" ]; then
    script_name="mongo"
  fi
  local script_file="${ROOT_DIR}/scripts/installation/${script_name}.sh"

  if [ ! -f "${script_file}" ]; then
    echo "Error: Installation script for service '${service_name}' not found at ${script_file}"
    return 1
  fi

  case "${action}" in
    start)
      echo "==> Starting ${service_name} service (script)..."
      # Source the script and call the install function
      if source "${script_file}" && declare -f "proj::${script_name}::docker::install" >/dev/null 2>&1; then
        "proj::${script_name}::docker::install"
      elif source "${script_file}" && declare -f "proj::${script_name}::install" >/dev/null 2>&1; then
        "proj::${script_name}::install"
      else
        echo "Warning: No install function found for ${service_name}, trying generic install"
        bash "${script_file}" install 2>/dev/null || echo "Failed to start ${service_name}"
      fi
      ;;
    stop)
      echo "==> Stopping ${service_name} service (script)..."
      # Source the script and call the uninstall function
      if source "${script_file}" && declare -f "proj::${script_name}::docker::uninstall" >/dev/null 2>&1; then
        "proj::${script_name}::docker::uninstall"
      elif source "${script_file}" && declare -f "proj::${script_name}::uninstall" >/dev/null 2>&1; then
        "proj::${script_name}::uninstall"
      else
        echo "Warning: No uninstall function found for ${service_name}, trying generic uninstall"
        bash "${script_file}" uninstall 2>/dev/null || echo "Failed to stop ${service_name}"
      fi
      ;;
    restart)
      echo "==> Restarting ${service_name} service (script)..."
      manage_script_service "${service_name}" "stop"
      sleep 2
      manage_script_service "${service_name}" "start"
      ;;
    status)
      echo "==> Checking ${service_name} service status (script)..."
      if source "${script_file}" && declare -f "proj::${script_name}::status" >/dev/null 2>&1; then
        "proj::${script_name}::status"
      else
        # Fallback: check docker container status
        local container_name="${NETWORK_NAME:-proj}-${service_name}"
        if docker ps --format "table {{.Names}}" | grep -q "${container_name}"; then
          echo "✅ ${service_name} container is running"
          docker ps --filter "name=${container_name}" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        else
          echo "❌ ${service_name} container is not running"
        fi
      fi
      ;;
    logs)
      echo "==> Showing ${service_name} service logs (script)..."
      # For script-based services, try to show docker logs if it's containerized
      local container_name="${NETWORK_NAME:-proj}-${service_name}"
      if docker ps --format "table {{.Names}}" | grep -q "${container_name}"; then
        docker logs -f "${container_name}"
      else
        echo "Log viewing for ${service_name} script-based service is not implemented"
        echo "Try: docker logs <container_name> or check system logs"
      fi
      ;;
    *)
      echo "Error: Invalid action '${action}' for script service"
      return 1
      ;;
  esac
}

# Function to manage a single service
manage_service() {
  local service_name=$1
  local action=$2
  
  # Check if service is supported
  if [[ ! -v SERVICE_TYPES["$service_name"] ]]; then
    echo "Error: Unsupported service '${service_name}'"
    return 1
  fi
  
  local service_type="${SERVICE_TYPES[$service_name]}"
  
  case "${service_type}" in
    "script")
      manage_script_service "${service_name}" "${action}"
      ;;
    *)
      echo "Error: Unknown service type '${service_type}' for service '${service_name}'"
      echo "All services should use 'script' type now."
      return 1
      ;;
  esac
}

# Function to handle service groups
handle_service_group() {
  local group=$1
  local action=$2
  local services=()
  
  case "${group}" in
    "all")
      services=("${ALL_SERVICES[@]}")
      ;;
    "database")
      services=("${DATABASE_SERVICES[@]}")
      ;;
    "observability") 
      services=("${OBSERVABILITY_SERVICES[@]}")
      ;;
    *)
      echo "Error: Unknown service group '${group}'"
      return 1
      ;;
  esac
  
  echo "==> Managing service group '${group}' with action '${action}'"
  local failed_services=()
  
  for service in "${services[@]}"; do
    echo "--- Processing ${service} ---"
    if ! manage_service "${service}" "${action}"; then
      failed_services+=("${service}")
      echo "Warning: Failed to ${action} service '${service}'"
    fi
    echo ""
  done
  
  if [ ${#failed_services[@]} -gt 0 ]; then
    echo "==> Some services failed: ${failed_services[*]}"
    return 1
  else
    echo "==> All services in group '${group}' processed successfully"
  fi
}

# Validate action
case "${ACTION}" in
  start|stop|restart|status|logs)
    ;;
  *)
    echo "Error: Invalid action '${ACTION}'"
    usage
    ;;
esac

# Main logic to handle the service and action  
case ${SERVICE} in
  # Individual services
  redis|kafka|jaeger|mariadb|mongodb|etcd|prometheus|grafana|alertmanager|otelcol|victorialogs)
    manage_service "${SERVICE}" "${ACTION}"
    ;;
  # Service groups
  all|database|observability)
    handle_service_group "${SERVICE}" "${ACTION}"
    ;;
  *)
    echo "Error: Invalid service '${SERVICE}'"
    echo ""
    echo "Available services: ${!SERVICE_TYPES[*]}"
    echo "Available groups: all, database, observability"
    usage
    ;;
esac

echo "==> Operation completed for ${SERVICE}."

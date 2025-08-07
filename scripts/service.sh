#!/usr/bin/env bash

# This script manages the lifecycle of dependent services using docker-compose.

set -o errexit
set -o nounset
set -o pipefail

# Define the root directory of the project
# This allows the script to be run from anywhere
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"

# Function to print usage
usage() {
  echo "Usage: $0 <action> <service>"
  echo ""
  echo "Actions:"
  echo "  start    Start the specified service."
  echo "  stop     Stop and remove the specified service."
  echo ""
  echo "Services:"
  echo "  jaeger   Jaeger tracing service."
  echo "  redis    Redis cache service."
  echo "  kafka    Kafka messaging service."
  echo "  all      Apply the action to all services."
  exit 1
}

# Check for the correct number of arguments
if [ "$#" -ne 2 ]; then
  usage
fi

ACTION=$1
SERVICE=$2

# Function to manage a single service
manage_service() {
  local service_name=$1
  local action=$2
  local compose_file="${ROOT_DIR}/deployments/${service_name}/docker-compose.yml"

  if [ ! -f "${compose_file}" ]; then
    echo "Error: docker-compose.yml for service '${service_name}' not found at ${compose_file}"
    exit 1
  fi

  # Capitalize first letter of action for prettier output
  local capitalized_action="$(tr '[:lower:]' '[:upper:]' <<< ${action:0:1})${action:1}"

  echo "==> ${capitalized_action}ing ${service_name} service..."
  if [ "${action}" == "start" ]; then
    docker-compose -f "${compose_file}" up -d
  elif [ "${action}" == "stop" ]; then
    docker-compose -f "${compose_file}" down
  else
    echo "Error: Invalid action '${action}'"
    usage
  fi
}

# Main logic to handle the service and action
case ${SERVICE} in
  jaeger|redis|kafka)
    manage_service "${SERVICE}" "${ACTION}"
    ;;
  all)
    for s in jaeger redis kafka; do
      manage_service "$s" "${ACTION}"
    done
    ;;
  *)
    echo "Error: Invalid service '${SERVICE}'"
    usage
    ;;
esac

echo "==> Operation completed successfully for ${SERVICE}."

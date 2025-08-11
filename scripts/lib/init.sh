#!/usr/bin/env bash


set -o errexit
set +o nounset
set -o pipefail

# Short-circuit if init.sh has already been sourced
[[ $(type -t proj::init::loaded) == function ]] && return 0


# Unset CDPATH so that path interpolation can work correctly
# https://github.com/minerrnetes/minerrnetes/issues/52255
unset CDPATH



# The root of the build/dist directory
PROJ_ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
SCRIPTS_DIR="${PROJ_ROOT_DIR}/scripts"

PROJ_OUTPUT_SUBPATH="${PROJ_OUTPUT_SUBPATH:-_output}"
PROJ_OUTPUT="${PROJ_ROOT_DIR}/${PROJ_OUTPUT_SUBPATH}"

source "${SCRIPTS_DIR}/lib/util.sh"
source "${SCRIPTS_DIR}/lib/logging.sh"
source "${SCRIPTS_DIR}/lib/color.sh"


proj::log::install_errexit

# Marker function to indicate init.sh has been fully sourced
proj::init::loaded() {
  return 0
}

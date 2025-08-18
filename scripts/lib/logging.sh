#!/usr/bin/env bash


# Controls verbosity of the script output and logging.
KUBE_VERBOSE="${KUBE_VERBOSE:-2}"


function proj::log::errexit() {
  local err="${PIPESTATUS[*]}"

  # If the shell we are in doesn't have errexit set (common in subshells) then
  # don't dump stacks.
  set +o | grep -qe "-o errexit" || return

  set +o xtrace
  local code="${1:-1}"
  # Print out the stack trace described by $function_stack
  if [ ${#FUNCNAME[@]} -gt 2 ]
  then
    proj::log::error "Call tree:"
    for ((i=1;i<${#FUNCNAME[@]}-1;i++))
    do
      proj::log::error " ${i}: ${BASH_SOURCE[${i}+1]}:${BASH_LINENO[${i}]} ${FUNCNAME[${i}]}(...)"
    done
  fi
  proj::log::error_exit "Error in ${BASH_SOURCE[1]}:${BASH_LINENO[0]}. '${BASH_COMMAND}' exited with status ${err}" "${1:-1}" 1
}


function proj::log::install_errexit() {
  # trap ERR to provide an error handler whenever a command exits nonzero  this
  # is a more verbose version of set -o errexit
  trap 'proj::log::errexit' ERR

  # setting errtrace allows our ERR trap handler to be propagated to functions,
  # expansions and subshells
  set -o errtrace
}

function proj::log::stack() {
  local stack_skip=${1:-0}
  stack_skip=$((stack_skip + 1))
  if [[ ${#FUNCNAME[@]} -gt ${stack_skip} ]]; then
    echo "Call stack:" >&2
    local i
    for ((i=1 ; i <= ${#FUNCNAME[@]} - stack_skip ; i++))
    do
      local frame_no=$((i - 1 + stack_skip))
      local source_file=${BASH_SOURCE[${frame_no}]}
      local source_lineno=${BASH_LINENO[$((frame_no - 1))]}
      local funcname=${FUNCNAME[${frame_no}]}
      echo "  ${i}: ${source_file}:${source_lineno} ${funcname}(...)" >&2
    done
  fi
}


# Log an error and exit.
# Args:
#   $1 Message to log with the error
#   $2 The error code to return
#   $3 The number of stack frames to skip when printing.
function proj::log::error_exit() {
  local message="${1:-}"
  local code="${2:-1}"
  local stack_skip="${3:-0}"
  stack_skip=$((stack_skip + 1))

  if [[ ${KUBE_VERBOSE} -ge 4 ]]; then
    local source_file=${BASH_SOURCE[${stack_skip}]}
    local source_line=${BASH_LINENO[$((stack_skip - 1))]}
    echo "!!! Error in ${source_file}:${source_line}" >&2
    [[ -z ${1-} ]] || {
      echo "  ${1}" >&2
    }

    proj::log::stack ${stack_skip}

    echo "Exiting with status ${code}" >&2
  fi

  exit "${code}"
}


# Log an error but keep going.  Don't dump the stack or exit.
function proj::log::error() {
  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "!!! ${timestamp} ${1-}" >&2
  shift
  for message; do
    echo "    ${message}" >&2
  done
}

# Print an usage message to stderr.  The arguments are printed directly.
function proj::log::usage() {
  echo >&2
  local message
  for message; do
    echo "${message}" >&2
  done
  echo >&2
}



function proj::log::usage_from_stdin() {
  local messages=()
  while read -r line; do
    messages+=("${line}")
  done

  proj::log::usage "${messages[@]}"
}

# Print out some info that isn't a top level status line
function proj::log::info() {
  local V="${V:-0}"
  if (( PROJ_VERBOSE < V )) || (( KUBE_VERBOSE < V )); then
    return
  fi

  for message; do
    echo "${message}"
  done
}

# Just like proj::log::info, but no \n, so you can make a progress bar
function proj::log::progress() {
  for message; do
    echo -e -n "${message}"
  done
}

function proj::log::info_from_stdin() {
  local messages=()
  while read -r line; do
    messages+=("${line}")
  done

  proj::log::info "${messages[@]}"
}

# Print a status line.  Formatted to show up in a stream of output.
function proj::log::status() {
  local V="${V:-0}"
  if (( KUBE_VERBOSE < V )); then
    return
  fi

  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "+++ ${timestamp} ${1}"
  shift
  for message; do
    echo "    ${message}"
  done
}

# Print a success message. Similar to status but for successful operations.
function proj::log::success() {
  local V="${V:-0}"
  if (( KUBE_VERBOSE < V )); then
    return
  fi

  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "✓✓✓ ${timestamp} ${1}"
  shift
  for message; do
    echo "    ${message}"
  done
}

# Print a warning message.
function proj::log::warn() {
  timestamp=$(date +"[%m%d %H:%M:%S]")
  echo "⚠⚠⚠ ${timestamp} ${1-}" >&2
  shift
  for message; do
    echo "    ${message}" >&2
  done
}

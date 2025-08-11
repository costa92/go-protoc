#!/usr/bin/env bash


function proj::util::sortable_date() {
  date "+%Y%m%d-%H%M%S"
}

# Run commands requiring root privileges without entering a password.
function proj::util::sudo()
{
  echo ${LINUX_PASSWORD} | sudo -S $1
}

# Run commands requiring root privileges without entering a password.
function proj::util::exec()
{
  eval "$@"
}


# Check if we're running on Linux
function proj::util::is_linux() {
    [[ "$(uname)" == "Linux" ]]
}

# Check if we're running on Mac
function proj::util::is_mac() {
    [[ "$(uname)" == "Darwin" ]]
}

# Check if a command exists
function proj::util::cmd_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Telnet is used to check if a port is up and running.
# $1: ip address, like: 127.0.0.1
# $2: port, like 3306
function proj::util::telnet()
{
  (
    set +o errexit
    set +o pipefail
    echo | telnet "$1" "$2" 2>&1|grep refused &>/dev/null
    if [ $? -eq 0 ]; then
      return 1
    fi
    return 0
  )
}

# ex: ts=2 sw=2 et filetype=sh

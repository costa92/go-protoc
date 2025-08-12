#!/usr/bin/env bash

# The root of the build/dist directory.
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${PROJ_ROOT_DIR}/scripts/installation/common.sh
# Set some environment variables.
PROJ_KAFKA_HOST=${PROJ_KAFKA_HOST:-127.0.0.1}
PROJ_KAFKA_PORT=${PROJ_KAFKA_PORT:-4317}
# 定义 zookeeper 和 kafka 的 docker 容器名称
ZOOKEEPER_DOCKER_MNAME=${NETWORK_NAME}-zookeeper
KAFKA_DOCKER_MNAME=${NETWORK_NAME}-kafka

# Install kafka using containerization.
# Refer to https://www.baeldung.com/ops/kafka-docker-setup
proj::kafka::docker::install()
{
  proj::common::network
  docker run -d --restart always --name ${ZOOKEEPER_DOCKER_MNAME} --network onex -p 2181:2181 -t wurstmeister/zookeeper
  docker run -d --name ${KAFKA_DOCKER_MNAME} --link ${ZOOKEEPER_DOCKER_MNAME}:zookeeper \
    --restart always \
    --network ${NETWORK_NAME} \
    --restart=always \
    -v /etc/localtime:/etc/localtime \
    -p ${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT}:9092 \
    --env KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181 \
    --env KAFKA_ADVERTISED_HOST_NAME=${PROJ_KAFKA_HOST} \
    --env KAFKA_ADVERTISED_PORT=${PROJ_KAFKA_PORT} \
    wurstmeister/kafka

  echo "Sleeping to wait for onex-kafka container to complete startup ..."
  sleep 5
  proj::kafka::status || return 1
  proj::kafka::info
  proj::log::info "install kafka successfully"
}

# Uninstall the docker container.
proj::kafka::docker::uninstall()
{
  docker rm -f ${ZOOKEEPER_DOCKER_MNAME} &>/dev/null
  docker rm -f ${KAFKA_DOCKER_MNAME} &>/dev/null
  proj::log::info "uninstall kafka successfully"
}

# Install the kafka step by step.
# sbs is the abbreviation for "step by step".
# Refer to https://kafka.apache.org/documentation/#quickstart
proj::kafka::install()
{
  proj::kafka::docker::install
  proj::log::info "install kafka successfully"
}

# Uninstall the kafka step by step.
proj::kafka::uninstall()
{
  proj::kafka::docker::uninstall
  proj::log::info "uninstall kafka successfully"
}

# Print necessary information after docker or sbs installation.
proj::kafka::info()
{
  echo -e ${C_GREEN}kafka has been installed, here are some useful information:${C_NORMAL}
  cat << EOF | sed 's/^/  /'
Kafka brokers is: ${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT}
EOF
}

# Status check after docker or sbs installation.
proj::kafka::status()
{
  proj::util::telnet ${PROJ_KAFKA_HOST} ${PROJ_KAFKA_PORT} || return 1
}

if [[ $* =~ proj::kafka:: ]]; then
  eval $*
fi
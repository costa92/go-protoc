#!/usr/bin/env bash

# The root of the build/dist directory.
PROJ_ROOT_DIR=$(dirname "${BASH_SOURCE[0]}")/../..
# If common.sh has already been sourced, it will not be sourced again here.
[[ -z ${COMMON_SOURCED} ]] && source ${PROJ_ROOT_DIR}/scripts/installation/common.sh
# Set some environment variables.
PROJ_KAFKA_HOST=${PROJ_KAFKA_HOST:-127.0.0.1}
PROJ_KAFKA_PORT=${PROJ_KAFKA_PORT:-9092}
# 定义 zookeeper 和 kafka 的 docker 容器名称
ZOOKEEPER_DOCKER_MNAME=${NETWORK_NAME}-zookeeper
KAFKA_DOCKER_MNAME=${NETWORK_NAME}-kafka

# Install kafka using containerization.
# Refer to https://www.baeldung.com/ops/kafka-docker-setup
proj::kafka::docker::install()
{
  proj::common::network

  # 检测系统架构
  local arch=$(uname -m)
  local platform="linux/amd64"
  if [[ "${arch}" == "aarch64" || "${arch}" == "arm64" ]]; then
    platform="linux/arm64"
  fi

  # 启动 Zookeeper (使用官方镜像，支持多架构)
  docker run -d --restart always --name ${ZOOKEEPER_DOCKER_MNAME} \
    --network ${NETWORK_NAME} \
    --platform ${platform} \
    -p 2181:2181 \
    -e ZOOKEEPER_CLIENT_PORT=2181 \
    -e ZOOKEEPER_TICK_TIME=2000 \
    confluentinc/cp-zookeeper:latest

  # 清理可能存在的同名容器
  proj::common::docker::cleanup_container "${KAFKA_DOCKER_MNAME}"

  # 启动 Kafka (使用支持ARM64的镜像)
  if [[ "${arch}" == "aarch64" || "${arch}" == "arm64" ]]; then
    proj::log::info "Install kafka arch: ${arch}"
    # ARM64架构使用bitnami镜像，原生支持ARM64，强制使用Zookeeper模式
    docker run -d --name ${KAFKA_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v /etc/localtime:/etc/localtime \
      -p ${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT}:9092 \
      -e KAFKA_ENABLE_KRAFT=no \
      -e KAFKA_CFG_ZOOKEEPER_CONNECT=${ZOOKEEPER_DOCKER_MNAME}:2181 \
      -e KAFKA_CFG_ADVERTISED_LISTENERS=PLAINTEXT://${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT} \
      -e KAFKA_CFG_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
      -e KAFKA_CFG_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
      -e ALLOW_PLAINTEXT_LISTENER=yes \
      bitnami/kafka:3.6
  else
    # x86_64架构使用confluent镜像
    docker run -d --name ${KAFKA_DOCKER_MNAME} \
      --restart always \
      --network ${NETWORK_NAME} \
      -v /etc/localtime:/etc/localtime \
      -p ${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT}:9092 \
      -e KAFKA_BROKER_ID=1 \
      -e KAFKA_ZOOKEEPER_CONNECT=${ZOOKEEPER_DOCKER_MNAME}:2181 \
      -e KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://${PROJ_KAFKA_HOST}:${PROJ_KAFKA_PORT} \
      -e KAFKA_LISTENERS=PLAINTEXT://0.0.0.0:9092 \
      -e KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR=1 \
      confluentinc/cp-kafka:6.2.0
  fi

  echo "Sleeping to wait for ${KAFKA_DOCKER_MNAME} container to complete startup ..."
  sleep 5
  proj::kafka::status || return 1
  proj::kafka::info
  proj::log::info "install platform ${platform} kafka successfully"
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
#!/usr/bin/env bash

# =============================================================================
# 重试机制库
# Retry Mechanism Library
# 
# 提供通用的重试机制、退避策略和失败处理功能
# =============================================================================

set -eEuo pipefail

# 避免重复加载
[[ "${RETRY_LOADED:-false}" == "true" ]] && return 0

# 默认重试配置
RETRY_MAX_ATTEMPTS="${RETRY_MAX_ATTEMPTS:-3}"
RETRY_DELAY="${RETRY_DELAY:-1}"
RETRY_MAX_DELAY="${RETRY_MAX_DELAY:-60}"
RETRY_BACKOFF_FACTOR="${RETRY_BACKOFF_FACTOR:-2}"

# =============================================================================
# 基础重试函数
# =============================================================================

# 简单重试函数
proj::retry::simple() {
    local max_attempts="${1:-${RETRY_MAX_ATTEMPTS}}"
    local delay="${2:-${RETRY_DELAY}}"
    shift 2
    local command=("$@")
    
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        if "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            proj::log::debug "Command failed, retrying in ${delay}s..."
            sleep "${delay}"
        else
            proj::log::error "Command failed after ${max_attempts} attempts"
        fi
        
        ((attempt++))
    done
    
    return 1
}

# 指数退避重试
proj::retry::exponential_backoff() {
    local max_attempts="${1:-${RETRY_MAX_ATTEMPTS}}"
    local initial_delay="${2:-${RETRY_DELAY}}"
    local backoff_factor="${3:-${RETRY_BACKOFF_FACTOR}}"
    local max_delay="${4:-${RETRY_MAX_DELAY}}"
    shift 4
    local command=("$@")
    
    local attempt=1
    local delay=$initial_delay
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        if "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            proj::log::debug "Command failed, retrying in ${delay}s..."
            sleep "${delay}"
            
            # 计算下次延迟时间
            delay=$((delay * backoff_factor))
            if [[ $delay -gt $max_delay ]]; then
                delay=$max_delay
            fi
        else
            proj::log::error "Command failed after ${max_attempts} attempts"
        fi
        
        ((attempt++))
    done
    
    return 1
}

# 线性退避重试
proj::retry::linear_backoff() {
    local max_attempts="${1:-${RETRY_MAX_ATTEMPTS}}"
    local initial_delay="${2:-${RETRY_DELAY}}"
    local increment="${3:-${RETRY_DELAY}}"
    local max_delay="${4:-${RETRY_MAX_DELAY}}"
    shift 4
    local command=("$@")
    
    local attempt=1
    local delay=$initial_delay
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        if "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            proj::log::debug "Command failed, retrying in ${delay}s..."
            sleep "${delay}"
            
            # 计算下次延迟时间
            delay=$((delay + increment))
            if [[ $delay -gt $max_delay ]]; then
                delay=$max_delay
            fi
        else
            proj::log::error "Command failed after ${max_attempts} attempts"
        fi
        
        ((attempt++))
    done
    
    return 1
}

# 随机延迟重试（防止雷群效应）
proj::retry::jitter() {
    local max_attempts="${1:-${RETRY_MAX_ATTEMPTS}}"
    local base_delay="${2:-${RETRY_DELAY}}"
    local max_jitter="${3:-$((base_delay / 2))}"
    shift 3
    local command=("$@")
    
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        if "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            # 添加随机延迟
            local jitter=$((RANDOM % (max_jitter * 2 + 1) - max_jitter))
            local delay=$((base_delay + jitter))
            
            # 确保延迟不为负数
            if [[ $delay -lt 0 ]]; then
                delay=0
            fi
            
            proj::log::debug "Command failed, retrying in ${delay}s..."
            sleep "${delay}"
        else
            proj::log::error "Command failed after ${max_attempts} attempts"
        fi
        
        ((attempt++))
    done
    
    return 1
}

# =============================================================================
# 条件重试函数
# =============================================================================

# 基于退出码的重试
proj::retry::on_exit_code() {
    local max_attempts="$1"
    local delay="$2"
    local exit_code="$3"
    shift 3
    local command=("$@")
    
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        if "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        local cmd_exit_code=$?
        
        if [[ $cmd_exit_code -eq $exit_code ]]; then
            if [[ $attempt -lt $max_attempts ]]; then
                proj::log::debug "Command failed with expected exit code (${exit_code}), retrying in ${delay}s..."
                sleep "${delay}"
            else
                proj::log::error "Command failed with exit code ${exit_code} after ${max_attempts} attempts"
            fi
        else
            proj::log::error "Command failed with unexpected exit code: ${cmd_exit_code}"
            return $cmd_exit_code
        fi
        
        ((attempt++))
    done
    
    return $exit_code
}

# 基于输出模式的重试
proj::retry::on_pattern() {
    local max_attempts="$1"
    local delay="$2"
    local pattern="$3"
    shift 3
    local command=("$@")
    
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        local output
        local exit_code
        
        output=$("${command[@]}" 2>&1) || exit_code=$?
        
        if [[ ${exit_code:-0} -eq 0 ]]; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            echo "${output}"
            return 0
        fi
        
        if echo "${output}" | grep -q "${pattern}"; then
            if [[ $attempt -lt $max_attempts ]]; then
                proj::log::debug "Command failed with expected pattern, retrying in ${delay}s..."
                sleep "${delay}"
            else
                proj::log::error "Command failed with pattern '${pattern}' after ${max_attempts} attempts"
                echo "${output}" >&2
            fi
        else
            proj::log::error "Command failed without expected pattern"
            echo "${output}" >&2
            return ${exit_code:-1}
        fi
        
        ((attempt++))
    done
    
    return 1
}

# =============================================================================
# 超时重试函数
# =============================================================================

# 带超时的重试
proj::retry::with_timeout() {
    local max_attempts="$1"
    local delay="$2"
    local timeout="$3"
    shift 3
    local command=("$@")
    
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts} with ${timeout}s timeout: ${command[*]}"
        
        if timeout "${timeout}" "${command[@]}"; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        local exit_code=$?
        
        if [[ $exit_code -eq 124 ]]; then
            # 超时错误
            if [[ $attempt -lt $max_attempts ]]; then
                proj::log::debug "Command timed out, retrying in ${delay}s..."
                sleep "${delay}"
            else
                proj::log::error "Command timed out after ${max_attempts} attempts"
            fi
        else
            # 其他错误
            if [[ $attempt -lt $max_attempts ]]; then
                proj::log::debug "Command failed (exit ${exit_code}), retrying in ${delay}s..."
                sleep "${delay}"
            else
                proj::log::error "Command failed with exit code ${exit_code} after ${max_attempts} attempts"
            fi
        fi
        
        ((attempt++))
    done
    
    return 1
}

# =============================================================================
# 高级重试函数
# =============================================================================

# 自适应重试（根据失败类型调整策略）
proj::retry::adaptive() {
    local max_attempts="${1:-${RETRY_MAX_ATTEMPTS}}"
    local base_delay="${2:-${RETRY_DELAY}}"
    shift 2
    local command=("$@")
    
    local attempt=1
    local consecutive_timeouts=0
    local consecutive_failures=0
    
    while [[ $attempt -le $max_attempts ]]; do
        proj::log::debug "Attempt ${attempt}/${max_attempts}: ${command[*]}"
        
        local start_time=$SECONDS
        local exit_code=0
        
        "${command[@]}" || exit_code=$?
        
        if [[ $exit_code -eq 0 ]]; then
            proj::log::debug "Command succeeded on attempt ${attempt}"
            return 0
        fi
        
        local duration=$((SECONDS - start_time))
        
        # 分析失败类型并调整策略
        local delay=$base_delay
        
        if [[ $exit_code -eq 124 ]]; then
            # 超时
            ((consecutive_timeouts++))
            delay=$((base_delay * (1 + consecutive_timeouts)))
            consecutive_failures=0
        else
            # 其他失败
            ((consecutive_failures++))
            delay=$((base_delay * RETRY_BACKOFF_FACTOR ** (consecutive_failures - 1)))
            consecutive_timeouts=0
        fi
        
        # 限制最大延迟
        if [[ $delay -gt $RETRY_MAX_DELAY ]]; then
            delay=$RETRY_MAX_DELAY
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            proj::log::debug "Command failed (exit ${exit_code}), retrying in ${delay}s... (timeouts: ${consecutive_timeouts}, failures: ${consecutive_failures})"
            sleep "${delay}"
        else
            proj::log::error "Command failed after ${max_attempts} attempts"
        fi
        
        ((attempt++))
    done
    
    return $exit_code
}

# =============================================================================
# 工具函数
# =============================================================================

# 重试配置验证
proj::retry::validate_config() {
    local max_attempts="$1"
    local delay="$2"
    
    if [[ ! $max_attempts =~ ^[0-9]+$ ]] || [[ $max_attempts -lt 1 ]]; then
        proj::log::error "Invalid max_attempts: ${max_attempts} (must be positive integer)"
        return 1
    fi
    
    if [[ ! $delay =~ ^[0-9]+(\.[0-9]+)?$ ]] || [[ $(echo "$delay < 0" | bc -l 2>/dev/null || echo "0") -eq 1 ]]; then
        proj::log::error "Invalid delay: ${delay} (must be non-negative number)"
        return 1
    fi
    
    return 0
}

# 获取推荐的重试策略
proj::retry::recommend_strategy() {
    local command_type="$1"
    
    case "${command_type}" in
        "network"|"http"|"api")
            echo "exponential_backoff with jitter"
            ;;
        "database"|"db")
            echo "exponential_backoff"
            ;;
        "file"|"io")
            echo "simple"
            ;;
        "docker"|"container")
            echo "linear_backoff"
            ;;
        *)
            echo "adaptive"
            ;;
    esac
}

# =============================================================================
# 批量重试函数
# =============================================================================

# 并行重试多个命令
proj::retry::parallel() {
    local max_attempts="$1"
    local delay="$2"
    shift 2
    local commands=("$@")
    
    local pids=()
    local results=()
    
    proj::log::info "Starting parallel retry for ${#commands[@]} commands"
    
    # 启动所有命令
    for i in "${!commands[@]}"; do
        (
            proj::retry::simple "$max_attempts" "$delay" bash -c "${commands[$i]}"
            echo $? > "/tmp/retry_result_$$_$i"
        ) &
        pids+=($!)
    done
    
    # 等待所有命令完成
    for i in "${!pids[@]}"; do
        wait "${pids[$i]}"
        if [[ -f "/tmp/retry_result_$$_$i" ]]; then
            results[i]=$(cat "/tmp/retry_result_$$_$i")
            rm -f "/tmp/retry_result_$$_$i"
        else
            results[i]=1
        fi
    done
    
    # 检查结果
    local failed_count=0
    for i in "${!results[@]}"; do
        if [[ ${results[$i]} -ne 0 ]]; then
            proj::log::error "Command $((i+1)) failed: ${commands[$i]}"
            ((failed_count++))
        else
            proj::log::debug "Command $((i+1)) succeeded: ${commands[$i]}"
        fi
    done
    
    if [[ $failed_count -gt 0 ]]; then
        proj::log::error "$failed_count out of ${#commands[@]} commands failed"
        return 1
    else
        proj::log::info "All commands succeeded"
        return 0
    fi
}

# =============================================================================
# 初始化函数
# =============================================================================

# 重试库初始化
proj::retry::init() {
    # 验证默认配置
    if ! proj::retry::validate_config "$RETRY_MAX_ATTEMPTS" "$RETRY_DELAY"; then
        proj::log::warn "Invalid default retry configuration, using fallback values"
        export RETRY_MAX_ATTEMPTS=3
        export RETRY_DELAY=1
    fi
    
    proj::log::debug "Retry library initialized (max_attempts: $RETRY_MAX_ATTEMPTS, delay: ${RETRY_DELAY}s)"
}

# 自动初始化
if [[ "${BASH_SOURCE[0]}" == "${0}" ]] || [[ "${RETRY_AUTO_INIT:-true}" == "true" ]]; then
    proj::retry::init
fi

# 标记已加载
export RETRY_LOADED=true
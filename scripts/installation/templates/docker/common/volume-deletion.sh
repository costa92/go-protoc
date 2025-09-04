#!/usr/bin/env bash

# Docker数据卷删除通用处理函数
# 使用方式: handle_volume_deletion "service_name" "volume1,volume2,..." "data_description"

handle_volume_deletion() {
    local service_name="$1"
    local volumes="$2"
    local data_description="$3"
    
    # 支持多种触发方式
    local should_remove_data=false
    
    # 1. 环境变量方式 (推荐)
    if [[ "${REMOVE_DATA:-}" == "true" ]]; then
        should_remove_data=true
    fi
    
    # 2. 命令行参数方式
    if [[ "${1:-}" == "--remove-data" ]] || [[ "${1:-}" == "--force" ]]; then
        should_remove_data=true
    fi
    
    # 3. 检查所有参数中是否有删除标志
    for arg in "$@"; do
        case $arg in
            --remove-data|--force)
                should_remove_data=true
                break
                ;;
        esac
    done
    
    if [ "$should_remove_data" = true ]; then
        echo ""
        echo "⚠️  警告：将删除${service_name}数据卷！这将永久删除所有数据！"
        echo "包括："
        echo "$data_description"
        echo ""
        
        # 交互式确认
        if [ -t 0 ]; then  # 检查是否为交互式终端
            read -p "确认删除数据卷？(y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                # 分割数据卷字符串并删除
                IFS=',' read -ra VOLUME_ARRAY <<< "$volumes"
                for volume in "${VOLUME_ARRAY[@]}"; do
                    volume=$(echo "$volume" | xargs)  # 去除空格
                    if docker volume rm "$volume" 2>/dev/null; then
                        echo "✅ 数据卷已删除: $volume"
                    else
                        echo "ℹ️  数据卷不存在或删除失败: $volume"
                    fi
                done
                echo "所有${service_name}数据已被永久删除"
            else
                echo "数据卷保留: $volumes"
            fi
        else
            # 非交互式环境，使用环境变量FORCE_DELETE
            if [[ "${FORCE_DELETE:-}" == "true" ]]; then
                IFS=',' read -ra VOLUME_ARRAY <<< "$volumes"
                for volume in "${VOLUME_ARRAY[@]}"; do
                    volume=$(echo "$volume" | xargs)
                    docker volume rm "$volume" 2>/dev/null || true
                done
                echo "数据卷已强制删除: $volumes"
            else
                echo "非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除"
                echo "数据卷保留: $volumes"
            fi
        fi
    else
        echo ""
        echo "💡 提示: ${service_name}数据卷已保留"
        IFS=',' read -ra VOLUME_ARRAY <<< "$volumes"
        for volume in "${VOLUME_ARRAY[@]}"; do
            volume=$(echo "$volume" | xargs)
            echo "   数据卷: $volume"
        done
        echo "   删除方式："
        echo "     环境变量: REMOVE_DATA=true make docker.${service_name,,}.stop"
        echo "     命令参数: bash <script> --remove-data"
    fi
}

# 导出函数供其他脚本使用
export -f handle_volume_deletion
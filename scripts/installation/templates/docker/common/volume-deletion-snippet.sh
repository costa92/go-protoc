# 通用数据卷删除处理 - 在停止脚本末尾添加此片段
# 需要定义以下变量：
# - SERVICE_NAME: 服务名称（用于显示）
# - DATA_VOLUMES: 数据卷列表，逗号分隔
# - DATA_DESCRIPTION: 数据描述，每行一个项目

# 支持多种数据删除触发方式
should_remove_data=false

# 方式1: 环境变量 (推荐用于Make命令)
if [[ "${REMOVE_DATA:-}" == "true" ]]; then
    should_remove_data=true
fi

# 方式2: 命令行参数
if [[ "${1:-}" == "--remove-data" ]] || [[ "${1:-}" == "--force" ]]; then
    should_remove_data=true
fi

# 方式3: 检查所有参数
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
    echo "⚠️  警告：将删除${SERVICE_NAME}数据卷！这将永久删除所有数据！"
    echo "包括："
    echo "$DATA_DESCRIPTION"
    echo ""
    
    # 交互式确认
    if [ -t 0 ]; then  # 检查是否为交互式终端
        read -p "确认删除数据卷？(y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            # 分割数据卷字符串并删除
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            deleted_volumes=()
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)  # 去除空格
                if docker volume rm "$volume" 2>/dev/null; then
                    deleted_volumes+=("$volume")
                fi
            done
            if [ ${#deleted_volumes[@]} -gt 0 ]; then
                echo "${SERVICE_NAME}数据卷已删除: ${deleted_volumes[*]}"
                echo "所有${SERVICE_NAME}数据已被永久删除"
            else
                echo "未找到要删除的数据卷"
            fi
        else
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    else
        # 非交互式环境，需要环境变量确认
        if [[ "${FORCE_DELETE:-}" == "true" ]]; then
            IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
            for volume in "${VOLUME_ARRAY[@]}"; do
                volume=$(echo "$volume" | xargs)
                docker volume rm "$volume" 2>/dev/null || true
            done
            echo "${SERVICE_NAME}数据卷已强制删除: $DATA_VOLUMES"
        else
            echo "非交互式环境，需要设置 FORCE_DELETE=true 进行强制删除"
            echo "数据卷保留: $DATA_VOLUMES"
        fi
    fi
else
    echo ""
    echo "💡 提示: ${SERVICE_NAME}数据卷已保留"
    IFS=',' read -ra VOLUME_ARRAY <<< "$DATA_VOLUMES"
    for volume in "${VOLUME_ARRAY[@]}"; do
        volume=$(echo "$volume" | xargs)
        echo "   数据卷: $volume"
    done
    echo ""
    echo "   删除数据的方式："
    echo "     环境变量: REMOVE_DATA=true make docker.${SERVICE_NAME,,}.stop"
    echo "     直接调用: bash <脚本路径> --remove-data"
fi
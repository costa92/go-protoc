# AWK 脚本

此目录包含 Makefile 系统用于生成帮助和目标信息的 AWK 脚本。

## 文件

### `help.awk`

- **用途**: 为 `make help` 命令生成格式化的帮助输出
- **架构**: 一体化处理器，单次处理所有 Makefile
- **功能**:
  - 解析多个文件中带有 `##` 注释的所有 Makefile 目标
  - 按类别分组目标（由 `##@` 标题定义）
  - 根据文件名动态转换内部目标名称（例如 `_install.*`）为用户友好的名称
  - 支持从 Makefile 名称动态提取前缀
  - 包含共享的颜色格式化常量和实用函数
  - 生成包含标题、目标和可选页脚的完整帮助输出

### `targets.awk`

- **用途**: 为 `make targets` 命令生成格式化的目标列表
- **架构**: 专为循环执行设计的单文件处理器
- **功能**:
  - 列出单个 Makefile 中的所有文档化目标
  - 应用与 `help.awk` 相同的动态前缀转换
  - 用于循环单独处理每个 Makefile
  - 通过内联公共函数与 `help.awk` 共享核心功能

## 实现设计

### 共享公共函数

两个脚本都包含相同的核心功能实现：

- **`get_file_prefix(filename)`**: 从 `.mk` 文件提取前缀（例如 `tools.mk` → `tools`）
- **`format_target(name, file_prefix)`**: 根据文件上下文应用命名规则
- **`process_category_header(line, indent_prefix)`**: 处理 `##@` 类别标题
- **`extract_var_target(target)`**: 处理基于变量的目标，如 `$(VAR)`
- **`extract_simple_target(target)`**: 处理简单目标，如 `build:`
- **`format_target_line(target_name, comment, indent_spaces)`**: 使用颜色格式化输出

### 目标名称转换规则

1. **特殊前缀**: `_install.`、`_uninstall.`、`_verify.` → 转换为 `{file_prefix}.install.` 等
2. **下划线转换**:
   - `.mk` 文件: `_` → `.` (点号记法)
   - 主 Makefile: `_` → `-` (连字符记法)
3. **变量目标**: `$(VAR)` → 提取变量名，转换为小写

### 颜色系统

两个脚本中一致的颜色方案：

- **青色** (`\033[36m`): 目标名称
- **品红** (`\033[35m`): 章节标题
- **粗体** (`\033[1m`): 类别名称
- **重置** (`\033[0m`): 清除格式

### 模式匹配

- **类别标题**: `/^##@/` - 匹配以 `##@` 开头的行
- **文档化目标**: `/^[^  #].*:[  ]*##/` - 匹配带有内联注释的目标定义
- **变量目标**: `/^\$\([a-zA-Z0-9._-]+\):.*?##/` - 匹配 `$(VAR): ## comment` 模式

## 动态前缀功能

两个脚本都支持基于 Makefile 名称的动态前缀生成：

- `tools.mk` → `_install.grpc` 变成 `tools.install.grpc`
- `service.mk` → 目标保持不变
- `database.mk` → `_install.migrate` 会变成 `database.install.migrate`

这允许灵活命名而无需硬编码特定的文件或类别名称。

## 使用方法

这些脚本自动被以下命令调用：

- `make help` - 使用 `help.awk`（一次性处理所有 Makefile）
- `make targets` - 使用 `targets.awk`（循环调用处理每个 Makefile）

## 语法要求

要使目标出现在帮助输出中，必须遵循以下格式：

```makefile
target-name: ## 此目标的作用描述
```

类别标题格式：

```makefile
##@ 类别名称
```

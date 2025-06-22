# 推荐命令

## 开发工具安装
```bash
make tools.install.*  # 安装所有开发工具
make install.ci      # 仅安装 CI 相关工具
```

## 代码生成
```bash
make proto           # 生成 Protocol Buffers 代码
make wire            # 生成依赖注入代码
```

## 开发命令
```bash
make fmt             # 格式化代码
make tidy            # 整理并更新依赖
make run-api         # 运行 API 服务器
make build           # 构建项目
```

## 代码检查
```bash
make apidiff         # 检查 API 变更
```

## 系统工具
```bash
# 文件操作
ls [-la]            # 列出文件和目录
cd <dir>            # 切换目录
grep <pattern>      # 搜索文件内容
find . -name        # 查找文件

# Git 操作
git status          # 查看仓库状态
git add .           # 暂存更改
git commit -m       # 提交更改
git push            # 推送更改
```
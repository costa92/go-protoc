# 任务完成检查清单

## 代码质量检查
1. 代码格式化
   ```bash
   make fmt
   ```

2. 依赖整理
   ```bash
   make tidy
   ```

3. API 兼容性检查
   ```bash
   make apidiff
   ```

## 生成代码更新
1. 更新 Protocol Buffers 代码
   ```bash
   make proto
   ```

2. 更新依赖注入代码
   ```bash
   make wire
   ```

## 测试验证
1. 运行单元测试
2. 本地运行服务验证
   ```bash
   make run-api
   ```

## 提交前检查
1. 确保所有生成的代码已更新
2. 确保配置文件格式正确
3. 确保日志输出合理
4. 确保错误处理完善
5. 确保文档已更新

## Git 操作
1. 添加修改的文件
2. 提交代码并写明清晰的提交信息
3. 推送到远程仓库
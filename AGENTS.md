# 仓库工作指南

GooseForum 使用 Go/Gin 后端和 resource/ 下的 React 主站及管理后台。

## 通用原则

- 保持改动简洁、职责清晰；优先沿用当前模块的模式，避免过度抽象或重复业务逻辑。
- 通过清晰的业务逻辑保证可靠性，不为断电保护引入恢复机制或持久化任务；只有确实需要保证业务数据一致性时才使用事务。
- 外部依赖优先选择稳定、主流、经过验证且契合当前技术栈的方案。
- 修改前阅读相关实现；文档与实现不一致时先核实，不将设计提案当成已实现功能。
- 按改动范围执行相关检查，说明未完成的验证。

## 专项说明

- 后端开发：app/AGENTS.md。
- 数据模型与查询：app/models/AGENTS.md。
- 前端、模板与资源构建：resource/AGENTS.md。
- 功能设计与运维说明：docs/；按当前任务查阅相关文档。

## 运行与配置

- 后端开发启动：`air`；直接运行：`./gooseforum serve`。
- Go 构建：`go build -ldflags="-w -s" .`；测试：`go test ./...`。
- 发布构建：`goreleaser build --snapshot --clean`，仅当前平台可加 `--single-target`。
- CLI：`./gooseforum set-user-admin <userId>`、`./gooseforum set-user-password <userId> <password>`。
- 运行配置入口为 config.toml；后台管理的站点设置也会存入数据库，不能假设全部配置都来自文件。
- 默认 HTTP 端口为 5234；app.debug 未配置时 local 默认开启，其他环境默认关闭。

# 后端开发

- http/controllers/、http/middleware/、http/routes/：请求转换、认证、响应和路由装配。
- service/：业务规则与流程协调。
- models/：数据模型和数据库访问，具体约定见 models/AGENTS.md。
- bundles/：配置、连接、分页、缓存等基础能力。
- migration/ 与 service/datamigration/：结构及业务数据迁移，沿用现有注册流程。
- console/：服务启动及 CLI。

数据库查询与更新封装到所属 model 包的 repository 方法，service 调用这些方法协调业务，不在 controller 中堆 SQL。需要跨步骤数据一致性时明确事务范围，不为普通多步骤操作默认加事务。

复用现有权限、账号限制、结构化消息和缓存失效入口；修改共享写入行为时检查浏览器、管理后台及 Agent API 等调用路径。

运行时配置先查 preferences；后台站点设置先查 pageConfig 和 hotdataserve，更新设置时沿用对应缓存失效方法。

验证按范围运行 `go test ./path/to/package -run TestName` 或 `go test ./...`；在仓库根目录构建 `go build -ldflags="-w -s" .`。数据库共享行为需考虑 SQLite、MySQL 和 PostgreSQL 的差异，不能把单一数据库的结果当成全部支持已验证。

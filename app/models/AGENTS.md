# 数据模型与数据库访问

本项目使用 GORM。连接由 app/bundles/connect/dbconnect 提供；业务模型主要位于 forum/，deprecatedmodels/ 是旧模型，新增功能先检查现行模型与迁移路径。

## 本地约定

- 模型结构及表名定义在对应模型文件；*_connect.go 中通常提供 `builder()`，底层为 `db.Connect().Table(tableName)`；查询和更新放 *_rep.go 或同包用途明确的 repository 文件。
- 优先复用模型包已有 builder、字段常量、queryopt 和 repository 方法，不另建一套 ORM 封装。标注“禁止手动更改”的文件不要直接编辑。
- 普通完整模型查询沿用现有 `builder().Where(...).First/Find(...)`。局部查询优先定义用途明确的小型读模型，使用 `builder().Model(&Entity{}).Where(...).First/Find(&readModel)`；用户完整模型名为 EntityComplete，其他包以实际定义为准。
- builder 已指定 Table，单独换成小接收结构不保证局部选列；需要推导列时显式提供完整 Model。参考 users.GetIdentity、GetAccountState 及 users_read_model_test.go。
- 不把禁止 Select 当成通用规则：普通字段投影优先小读模型；单列可用 Pluck；存在性、聚合、连接别名或计算表达式按需求使用 Select/Scan。参数值保持绑定，不拼接外部输入。
- 更新按业务意图限定字段。需要写入 false、0 或空字符串时注意 GORM 结构体 Updates 忽略零值的行为，可沿用现有字段 map。账号敏感字段使用专用更新入口，不能被通用 Save 覆盖。
- 使用 Table、Model、Raw 或连接查询时检查软删除语义、限定条件与实际 SQL，不假设所有写法都会自动过滤软删除。
- repository 返回业务所需的 error/RowsAffected，由 service 判断不存在、冲突与后续流程；新查询不要复制旧代码中忽略错误的做法。

## 兼容与验证

支持 SQLite、MySQL 和 PostgreSQL。默认 SQLite 文件位于 storage/database/sqlite.db；连接配置和实际路径以当前配置/连接实现为准。

涉及局部字段、软删除、关联或条件更新时，优先通过现有 SQL 断言/数据库测试验证真实行为。涉及结构或数据变更时检查 migration/ 和 service/datamigration/，不能只改 Go struct。

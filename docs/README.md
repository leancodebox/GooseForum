# GooseForum 文档目录

这里收拢 GooseForum 的长期文档、设计记录和归档资料。根目录 README 保持面向新用户，细节文档放在本目录下维护。

## 用户与部署

- [配置文档](user/configuration.md)：`config.toml` 配置项、数据库、日志和常见问题。

## 架构与开发

- [Resource 前端架构](architecture/resource-frontend.md)：`resource/` 前端结构、渲染模型和 payload 约定。
- [Markdown 渲染方向](architecture/markdown-rendering.md)：Markdown 服务端/客户端双实现、兼容测试和增强渲染边界。
- [主题访问控制](architecture/topic-access-control.md)：Issue #12 的权限模型、查询边界、迁移与渐进交付设计。
- [访问控制性能基线](architecture/topic-access-control-performance.md)：10 万主题的 SQLite 可见率基准与复现方法。
- [Resource UI 规范](frontend/ui-spec.md)：前端 UI 规则、布局约束和组件风格。

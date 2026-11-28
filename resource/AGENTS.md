# 前端与模板开发

技术栈为 React 19、TypeScript、Vite、TailwindCSS 4，使用 pnpm workspace。

## 目录职责

- apps/web/src/site/：主站启动、导航、缓存和主题装配。
- apps/web/src/host/：Go/Vite 与浏览器环境适配。
- packages/theme-default/src/site/：默认主题页面和业务组件。
- apps/web/src/admin/：管理后台入口、布局、页面和样式。
- packages/runtime/src/：运行时接口、Context/hooks 和主站国际化接入。
- packages/i18n/：前端词条目录、生成词典的运行时和类型。词条的增删改、翻译和审核均使用 pnpm i18n 命令。用法见 [前端国际化](packages/i18n/README.md)。
- packages/ui/src/：共享基础组件、工具和样式。
- packages/client/src/contracts/：共享接口类型。
- templates/：首屏、SEO 和 no-js 降级的 GoHTML 模板。

改动遵循现有包职责和依赖边界，业务权限由后端校验，前端使用后端能力和结构化错误展示状态。

## 开发与构建

修改此目录中的模板、样式或脚本后，必须按仓库 config.toml 的当前 app.env 处理资源：

- local：在 resource/ 运行 `pnpm dev`。
- production：在 resource/ 运行 `pnpm build`。
- 其他环境：先核实实际资源加载模式，再选择开发服务或生产构建。

页面样式、资源引用或模板异常时，先检查开发服务或构建产物是否更新。

## 验证

以下命令均在 resource/ 执行，按改动范围选择：

- 类型检查：`pnpm typecheck`。
- 前端测试：`pnpm test`。
- 国际化检查：`pnpm check:i18n` 校验词条、ICU 参数、语言完整性和 API 消息码。用 `pnpm i18n status` 查看各语言的缺失、过期和待审核词条。
- 包依赖边界：`pnpm check:boundaries`。
- 完整 React 检查：`pnpm check:react`。
- 浏览器端到端测试：`pnpm test:react:e2e`。

脚本定义以 package.json 为准；修改实际界面后检查交互与响应式表现，不能只依赖构建通过。

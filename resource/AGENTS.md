# 前端与模板开发

技术栈为 React 19、TypeScript、Vite、TailwindCSS 4，使用 pnpm workspace。

## 目录职责

- apps/web/src/site/：主站启动、导航、缓存和主题装配。
- apps/web/src/host/：Go/Vite 与浏览器环境适配。
- packages/theme-default/src/site/：默认主题页面和业务组件。
- apps/web/src/admin/：管理后台入口、布局、页面和样式。
- packages/runtime/src/：运行时接口、Context/hooks 和国际化。
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
- 国际化审计：`pnpm check:i18n` 校验各语言实际加载的键、占位符和已识别的静态引用。
- 审查词典改动：`pnpm audit:i18n --base HEAD --history` 比较工作区和指定提交，列出新增中文键、调用位置及基线调用文件中的同文案。历史匹配限同文件、同键和基线的第一父链；检查 PR 时将 HEAD 换为合并基线。用 `pnpm audit:i18n --all --json` 导出逐键来源和动态候选；未识别引用的键需要复核后再删除。报告中的 `lostReferenceCandidates` 优先列出基线有调用线索、本次改动后只剩词典定义的候选，并区分明确键与动态前缀。
- 批量筛选未识别键：在 resource/ 执行 `pnpm audit:i18n --groups` 查看按线索和命名空间汇总的数量，追加 `--json` 导出每个键的全部线索和位置。默认审计和分组只读取当前工作区。优先复核 `unusedCandidate`：已排除已识别的调用、动态候选和 API 契约，再检查各语言的共享定义、前端同名字符串、属性和动态前缀，以及 Go 键名字面量。候选只代表当前扫描范围内缺少使用线索。
- 包依赖边界：`pnpm check:boundaries`。
- 完整 React 检查：`pnpm check:react`。
- 浏览器端到端测试：`pnpm test:react:e2e`。

脚本定义以 package.json 为准；修改实际界面后检查交互与响应式表现，不能只依赖构建通过。

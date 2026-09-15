# C 端 shadcn-vue 迁移章程

> 历史记录：Vue 实现已退役。当前实现位于 resource/apps/web 与 resource/packages/theme-default；本文中的旧路径仅供 Git 历史对照，不再是开发规范。

## 目标

以 shadcn-vue 官方组件源码和 Reka UI primitives 作为 C 端基础 UI 层，保留 GooseForum 现有的信息架构、内容密度、主题能力和产品辨识度。迁移完成后，不再维护与官方组件能力重复的手写基础交互。

## 执行原则

1. **官方优先**：需要基础组件时，先检查 shadcn-vue 官方 registry；存在时必须通过 `pnpm dlx shadcn-vue@latest add <component>` 下载，不凭记忆手写，也不从管理端复制。
2. **源码入库**：下载后的组件是 GooseForum 拥有的源码。适配现有 token、尺寸与交互规范，但保留组件结构、无障碍语义和 props/emits 转发能力。
3. **视觉等价优先**：第一阶段迁移不改变页面布局、信息层级与内容密度。允许的视觉变化必须能说明具体收益，例如焦点可见性、触控面积或状态辨识度。
4. **产品组件不通用化**：Topic、Post、Composer、UserCard、AppShell 等继续作为论坛产品组件，只在内部组合基础 UI，不降级为通用 Card/Button 的堆叠。
5. **纵向切片**：一次迁移一个可完整验证的页面或交互，不长期保留新旧实现并行的适配层。
6. **行为先于装饰**：优先替换 Select、Dialog、Drawer、DropdownMenu、Popover、Tooltip 等容易产生键盘、焦点和浮层缺陷的交互。
7. **主题只有一个来源**：`src/styles/tokens.css` 是品牌和用户主题 token 的事实来源；shadcn 语义变量只能映射这些 token，不另建竞争性的主题配置。
8. **下载后审查**：每次 CLI 下载后检查新增依赖、生成源码和 CSS 变更。社区 registry 不能自动采用；确需使用时必须先审查来源与代码。
9. **升级不覆盖定制**：不使用 `--overwrite` 批量覆盖已经适配的组件。上游升级通过 dry-run/diff 逐个合并，并记录有意义的行为变化。
10. **删除才算完成**：调用方迁移并验证后，删除对应旧组件、旧事件监听和旧 CSS；仅新增 shadcn 文件不算完成迁移。

## 目录边界

```text
resource/src/components/ui/       Admin 与 C 端共享的 shadcn-vue 基础 UI 源码
resource/src/lib/utils.ts         共享 class 合并工具
resource/src/site/components/     C 端论坛产品组件
resource/src/admin/components/    管理端产品组件
resource/src/site/pages/          页面编排
resource/src/styles/tokens.css    主题 token 唯一来源
```

Admin 与 C 端必须从 `src/components/ui` 导入基础组件，不得在各自目录建立第二份 shadcn-vue 源码。两端不同的页面密度和产品表达通过 variants、调用方 class 与业务组件组合实现。

## 迁移顺序

### 阶段 1：基础设施与样板

- 配置 `components.json`，确保官方 CLI 始终写入共享组件目录。
- 将管理端现有 shadcn-vue 组件提升为共享基础层，并迁移管理端引用。
- 下载并适配 Button、Input、Textarea、Label、Badge、Avatar、Skeleton、Separator。
- 在 ThemePreviewPage 建立明暗主题、状态和尺寸样板。

### 阶段 2：浮层与选择交互

- Select 替换 SiteSelect。
- DropdownMenu 替换 AppShell 菜单。
- Tooltip、Popover 替换页面内手写浮层。
- Dialog、AlertDialog 替换确认框和表单弹窗。
- Drawer/Sheet 替换 MobileDrawer。

### 阶段 3：页面纵向迁移

1. Login、ResetPassword、OIDCConsent
2. Search、Categories、Members、Links、Sponsors
3. Settings、ThemePreview
4. Publish、Drafts
5. Topic、PostComposer、图片查看器
6. Messages、Notifications、Moderation
7. Home、Category、AppShell 收尾

### 阶段 4：清理与收敛

- 删除已失去调用方的 `gf-button`、`gf-input`、`gf-textarea` 等旧基础样式。
- 保留 `gf-card`、`gf-panel`、排版、编辑器和论坛布局等产品样式。
- 删除重复交互状态机和全局事件监听。
- 检查基础组件 variants，删除只为单一调用方遗留且没有产品含义的分支。

## 每个切片的完成标准

- 组件来自官方 registry，或明确记录官方不存在。
- 默认、hover、focus-visible、active、disabled、invalid、loading 状态完整。
- 鼠标、触摸和键盘均可完成操作；弹层焦点进入、循环和恢复正确。
- 明亮、暗色和自定义主题表现正确。
- 手机与桌面布局无回归。
- `pnpm typecheck`、相关 Vitest 和前端构建通过。
- 旧调用、旧逻辑和无用样式已经删除。

## 组件台账

| 组件组 | 当前状态 | 目标 | 迁移阶段 |
| --- | --- | --- | --- |
| Button/Input/Textarea | 调用方与旧 CSS 已清理 | 保持共享实现 | 1（完成） |
| Badge | 调用方与旧 CSS 已清理 | 保持共享实现 | 1（完成） |
| Avatar | UserAvatar 已组合官方 Avatar/Fallback | 保持产品组件包装 | 1（完成） |
| Label/Skeleton/Separator | 无 C 端基础调用方，保留原生产品语义 | 按需使用共享实现 | 1（完成） |
| SiteSelect | 已迁移至共享 shadcn Select | 保持共享实现 | 2（完成） |
| AppShell menus | 已迁移至共享 shadcn DropdownMenu | 保持共享实现 | 2（完成） |
| Popover | C 端手写浮层已迁移 | 保持共享实现 | 2（完成） |
| Tooltip | C 端手写浮层已迁移 | 保持共享实现 | 2（完成） |
| Tabs | C 端 segmented/tab、语言切换已迁移 | 保持共享实现 | 2（完成） |
| 页面确认框 | 手写 Teleport/Dialog | shadcn AlertDialog/Dialog | 2-3（完成） |
| MobileDrawer | 已迁移至共享 shadcn Sheet | 保持共享实现 | 2（完成） |
| Topic/Post/Composer | 产品组件 | 保持产品组件，内部组合 primitives | 3（PostComposer 完成） |

## 完成状态

C 端基础 UI 迁移已完成。`pnpm scan:site-dom` 的当前基线为：

- 原生 action `<button>`：0
- 手写 `role="dialog"` / `aria-modal` 浮层：0
- 旧基础控件调用（`gf-button`、`gf-input`、`gf-textarea`、`gf-tab`、`gf-segmented`、`gf-menu-surface` 等）：0
- Tab 高度挤压风险：0
- Button 缺失 variant 风险：0

保留的原生 `<a>`、文件/滑杆输入和专用编辑器 textarea 属于论坛产品语义或专用编辑器实现，不是与 shadcn-vue 基础组件重复的手写交互。后续改动应运行 `pnpm scan:site-dom`，并保持上述基线不回退。

`scan:site-dom` 默认作为守门检查运行：发现原生 action `<button>`、旧基础控件类、手写 dialog、Tab 高度风险或 Button 缺失 variant 时返回失败。浮动定位元素与全局事件监听只保留在清单中供 CR 使用，不直接判定失败。

admin 与 C 端共用唯一的 `resource/src/components/ui`。`pnpm scan:admin-dom` 会阻止 admin 重新引入原生 button/select/textarea、普通原生 input、旧控件类或手写 dialog；文件上传 input 保留。`pnpm scan:dom` 可一次检查两个入口。

admin 基础 UI 同样完成收口：34 个 SFC 中原生 action button、select、textarea、普通 input、旧控件类与手写 dialog 均为 0；仅保留 3 个浏览器文件选择 input。顶部语言菜单和管理员用户搜索候选框分别使用共享 DropdownMenu、Popover 与 Command。

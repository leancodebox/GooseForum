# Resource V2 前端迁移设计

## 目标

本文档定义 GooseForum 新一代前台资源层的迁移方案。目标不是在现有模板里继续补丁式改造，而是在保持 Go 单体应用、SEO 稳定、首屏不闪烁的前提下，建立一套更清晰、可渐进替换、最终可彻底切换的新前端架构。

核心目标：

- 保持 Go 单体应用，不引入独立 Node SSR 服务。
- 首屏由服务端输出可用 HTML，避免白屏和明显闪烁。
- 搜索引擎和无 JavaScript 环境可以读取核心公开内容。
- 首屏之后提供接近 SPA 的站内跳转体验。
- 新实现放在独立边界内，迁移完成后可以一次性切换并删除旧实现。
- 功能覆盖老前台，不要求像素级复刻老 UI。
- 视觉和交互以 Discourse 为主要参考，同时沉淀 GooseForum 自己的 UI 规范。

## 非目标

- 不追求把老页面 100% 像素级复制到新页面。
- 不引入 Vite 作为必须的运行时服务。
- 不把 SEO 依赖建立在公开的 `?render=seo` 查询参数上。
- 不把 Go 模板变量直接拼进复杂 JavaScript 字符串。
- 不在 Go 代码中维护前端选择器、root margin、Tailwind class 等样式细节。

## 总体方案

Resource V2 采用“服务端首屏 HTML + 前端增强”的渐进式架构。

普通用户访问页面时，Go 服务端返回完整页面外壳、SEO meta、核心 noscript 内容、页面初始化 JSON，以及前端入口脚本。JavaScript 加载后接管站内导航、交互和局部刷新。搜索引擎或禁用 JavaScript 的访问者仍然可以从服务端 HTML 中读取公开内容。

这不是传统纯 SPA，也不是独立 SSR。它更接近 Discourse、Flarum 这类论坛产品的思路：服务端保证可索引、可退化，前端负责增强体验。

## 目录边界

建议新增独立目录：

```text
resource/
  templates/
    layout/
    pages/
    partials/
    seo/
  src/
    app/
    pages/
    components/
    types/
    main.ts
```

旧前台继续保留在 `resource/`。迁移阶段通过新路由前缀访问 V2，避免新旧实现互相污染。

原则：

- V2 模板不继续堆在 `resource/templates/view`。
- V2 可以复用现有静态资源、Tailwind 构建能力和部分样式经验。
- V2 不直接依赖旧模板 partial，避免最终删除旧实现时留下尾巴。
- 如果确实要借鉴旧 layout，应复制并整理成 V2 自己的 layout，而不是跨目录 include。

## 模板策略

V2 模板建议拆成几类：

```text
resource/templates/layout/app.gohtml
resource/templates/layout/head.gohtml
resource/templates/layout/sidebar.gohtml
resource/templates/pages/shell.gohtml
resource/templates/seo/home.gohtml
resource/templates/seo/category.gohtml
resource/templates/seo/article.gohtml
resource/templates/partials/topic_row.gohtml
resource/templates/partials/post.gohtml
```

模板职责：

- `layout`：页面骨架、head、全局导航、侧边栏、noscript 容器。
- `pages/shell`：前端挂载点和初始化数据。
- `seo`：公开页面的服务端 HTML 内容。
- `partials`：SEO/no-js 与测试可复用的内容片段。

Go 模板只负责输出结构化数据和可退化 HTML，不负责前端交互细节。

## 渲染模式

同一个公开 URL 应支持三种访问形态：

1. 普通首屏 HTML。
2. 前端站内跳转请求的 JSON payload。
3. 搜索引擎或无 JavaScript 场景下的服务端可读内容。

普通首屏响应包含：

- 标准 HTML 文档。
- SEO meta、canonical、Open Graph 等 head 信息。
- V2 app mount point。
- `script[type="application/json"]` 初始化数据。
- `noscript` 或服务端可读内容区。
- 前端入口脚本。

站内跳转请求通过请求头区分，例如：

```text
X-Goose-Page: true
```

返回结构化 JSON，而不是整页 HTML。

不建议把 `?render=seo` 作为公开协议。搜索引擎不会主动传这个参数，而且公开查询参数会制造额外 canonical、缓存和误用问题。

当前已验收实现采用统一首屏 shell：普通 HTML 响应包含 head meta、初始化 payload、前端入口和 `<noscript>` fallback。公开页面的核心可读内容放在 `<noscript>` 中，并与 JSON payload 共用服务端 DTO。后续如需增加 crawler 专用响应，应继续复用同一 DTO 和 partial，不应引入公开查询参数或另一套内容来源。

调试可以使用内部手段：

- Go 测试直接调用渲染函数。
- 检查普通 HTML 中的 `<noscript>` fallback。
- 使用 `X-Goose-Page: true` 检查 JSON payload。

## Page Payload

前端接收统一的页面数据：

```go
type PagePayload struct {
    Component string        `json:"component"`
    Props     any           `json:"props"`
    Meta      PageMeta      `json:"meta"`
    Layout    LayoutPayload `json:"layout"`
    URL       string        `json:"url"`
    Version   string        `json:"version"`
}
```

约束：

- `Component` 只描述页面类型，例如 `article.detail`、`topic.list`。
- `Props` 只放业务数据。
- `Meta` 只放 title、description、canonical、robots 等 SEO 信息。
- `Layout` 放站点信息、登录用户、侧边栏数据、当前导航状态。
- 不在 payload 中传 DOM selector、CSS class、root margin、动画参数等前端实现细节。
- UI 文案不应每页重复下发，通用文案由前端 i18n 或共享字典维护。

## Layout Payload

建议结构：

```go
type LayoutPayload struct {
    Site      SitePayload      `json:"site"`
    Viewer    ViewerPayload    `json:"viewer"`
    Sidebar   SidebarPayload   `json:"sidebar"`
    Footer    FooterPayload    `json:"footer"`
}
```

侧边栏数据应由服务端提供信息架构，由前端决定具体表现：

```go
type SidebarPayload struct {
    Main       []NavItemPayload      `json:"main"`
    Resources  []NavItemPayload      `json:"resources"`
    Categories []CategoryNavPayload  `json:"categories"`
    ActiveKey  string                `json:"activeKey"`
}
```

这样可以避免每个页面重复实现侧边栏，也避免把导航激活逻辑散落在模板和脚本里。

## SEO 与 No-JS

公开页面必须有服务端可读内容。这里的“SEO 稳定性来自服务端 fallback”指的是：即使 JavaScript 不执行，页面 HTML 里仍然存在核心内容、链接和 meta，而不是只剩一个空的 `div#app`。

必须支持 no-js 的页面：

- 首页主题列表。
- 分类主题列表。
- 文章详情。
- 用户公开主页。
- 搜索结果页。
- 友情链接、赞助、关于等公开信息页。
- 登录/注册的基础表单可访问。

可以只提供有限 no-js 能力的页面：

- 通知。
- 私信。
- 设置。
- 发布/编辑。

这些页面至少不能空白，应给出可理解的服务端说明或基础表单。

no-js 内容和 crawler 内容应尽量共用模板 partial，避免普通用户、搜索引擎、禁用 JS 用户看到三套不同内容。

## 站内 SPA 体验

JavaScript 启动后，应拦截同源 V2 链接并走 JSON payload。

拦截条件：

- 同源链接。
- 属于 V2 管理范围。
- 非下载链接。
- 非 `target="_blank"`。
- 用户没有按住 meta、ctrl、shift、alt 等修饰键。

跳转流程：

1. 拦截链接点击。
2. 请求目标 URL，带上 `X-Goose-Page: true`。
3. 服务端返回 `PagePayload`。
4. 前端更新当前组件、props、layout、title、meta、canonical、侧边栏 active。
5. 调用 `history.pushState`。
6. 恢复滚动位置或滚动到顶部。

失败策略：

- JSON 请求失败时退回浏览器原生跳转。
- 401/403 根据登录状态和权限展示对应页面或跳转。
- 后退/前进使用 `popstate` 恢复页面。
- 表单默认不做全局拦截，除非该表单已经明确实现前端增强。

## UI 设计方向

Discourse 是 V2 的主要视觉和交互参考。老前台也是参考材料，但它不是设计规范本身。

V2 不追求老 UI 像素级一致，而追求：

- 功能 100% 覆盖。
- 信息架构完整。
- 核心交互不弱于老页面。
- 视觉密度、阅读体验、论坛产品感接近 Discourse。
- 有明确、可复用、可检查的 UI 规范。

UI 细节以 `RESOURCE_V2_UI_SPEC.md` 为准。实现过程中如果发现规范不够，应先补规范，再落代码。

## 路由迁移策略

迁移阶段使用临时前缀：

```text
/v2
```

第一批页面：

```text
/
/c/:slug/:id
/c/:slug/:id/l/:sort
/p/post/:id
```

第二批页面：

```text
/search
/links
/sponsors
/u/:userId
```

第三批页面：

```text
/notifications
/settings
/messages
```

第四批页面：

```text
/publish
/publish?id=:id
```

切换步骤：

1. V2 所有页面通过功能、SEO、no-js、交互和 UI 验收。
2. 正式路由指向 V2。
3. 短期保留旧实现作为回退。
4. 确认稳定后删除旧前台模板、旧页面脚本和相关 Alpine handler。

## Go 侧结构

建议新增控制器包：

```text
app/http/controllers/
  page.go
  layout.go
  home.go
  category.go
  article.go
  search.go
```

职责：

- 页面控制器负责查询业务数据。
- DTO 组装层负责转换为前端 payload。
- `renderPage` 统一处理 HTML、JSON、SEO/no-js 渲染。
- layout 数据统一构建，避免每个控制器重复拼导航。

DTO 规则：

- 不直接暴露 GORM model。
- 不把模板专用字段混入业务 model。
- 不把 UI 样式配置放进 Go DTO。
- URL、权限、计数、时间展示所需原始字段应明确提供。

## 模板注册

模板注册需要同时加载旧模板和 V2 模板：

```text
resource/templates/**/*.gohtml
```

V2 已经处在独立 `resource` 目录内，文件名不再重复带 `v2` 前缀。模板命名按职责分目录，例如：

```text
layout/app.gohtml
pages/home.gohtml
partials/topic_list.gohtml
seo/article.gohtml
```

避免与旧模板同名。

## 构建策略

V2 使用独立前端工程，不复用旧 `resource` 的 Alpine 入口和页面脚本：

```text
resource/
  package.json
  vite.config.ts
  src/
    main.ts
    App.vue
    pages/
    runtime/
    types/
    styles/
```

首批采用：

```text
Vue 3 + TypeScript + Vite + Tailwind CSS 4
```

选择原因：

- Vue 3 稳定、生态成熟，适合从服务端 HTML 渐进增强为 SPA。
- TypeScript 用于约束 PagePayload 和页面 props，避免模板、Go DTO、前端组件各写一套隐式结构。
- Vite 继续作为构建工具，但 V2 dev server 使用独立端口，默认 `3010`。
- Tailwind CSS 4 继续服务 UI token 和模板/组件样式。

构建产物输出到：

```text
resource/static/dist
```

Go 侧通过 `ResourceEntry` 读取 `resource/static/dist/.vite/manifest.json`，生产环境静态资源通过 `/assets/` 暴露。开发环境 `ResourceEntry` 指向 `http://localhost:3010/src/main.ts`。

Tailwind content/source 必须覆盖 V2 模板和源码，否则样式会被裁剪：

```text
resource/templates/**/*.gohtml
resource/src/**/*.{ts,vue}
```

V2 不应导入旧 `resource/src/gf-main.ts`，也不应依赖旧 Alpine handler。

### Resource V2 构建命令

```bash
cd resource && pnpm dev
cd resource && pnpm build
```

后续如果需要一体化启动，应在 `dev.sh` 中显式加入 `resource` dev server，而不是借用旧 `resource` dev server。

## 前端运行时规则

V2 前端使用共享 App Shell，不允许每个页面重复实现 Header、Sidebar、移动端 Drawer 和站内跳转加载状态。

当前共享层：

```text
resource/src/components/AppShell.vue
resource/src/runtime/router.ts
resource/src/runtime/navigation-state.ts
resource/src/runtime/format.ts
```

规则：

- 页面组件只负责自己的主体内容和页面专属 right rail。
- Header、Sidebar、移动端 Drawer、登录/发布入口由 App Shell 统一渲染。
- 站内跳转必须显示轻量加载条，旧内容保留到新 payload 到达。
- 数据变更成功后优先重新请求当前 `PagePayload` 并局部更新页面，不用整页刷新。
- 格式化逻辑如数字压缩、时间显示放在 runtime helper，不散落在组件里。
- Header 工具按钮使用 `@lucide/vue` 图标，避免临时字符或手写 SVG 扩散。

## UI Spec 对齐状态

当前 App Shell 已按 `RESOURCE_V2_UI_SPEC.md` 落地以下规则：

- Header 高度 64px，sticky，包含品牌、搜索入口、语言切换、通知入口、用户菜单、登录/注册、发布入口和移动端菜单按钮。
- Sidebar 使用 `lg 210px / xl 224px`，桌面端 sticky，移动端进入 drawer。
- Sidebar 行高和分区按高密度论坛导航处理，使用 13px 文本、紧凑 padding、轻量 active 状态。
- 首页主题列表改为紧凑 topic stream，突出标题、作者、分类、摘要、回复/浏览/活跃时间，不把每行做成大卡片。
- 详情页改为阅读流结构，原帖与回复使用一致的 post stream 语言，right rail 只承载主题信息和回复地图。
- 页面主体最大宽度限制为 1600px，详情页在 `xl` 以上提供 right rail。
- 页面结构以边框为主，阴影保持克制。

仍需补齐：

- 搜索入口目前是页面链接，尚未实现弹层或快速搜索。
- 通知入口目前只有入口，没有未读状态和预览菜单。
- 用户菜单已有基础项，但还未补头像加载失败、键盘导航和 focus trap。
- 移动端 drawer 已有基础关闭行为，但还未实现严格 focus trap 和 body scroll lock。

## 首页页面契约

第一条落地页面是 `/`。

### 路由矩阵

```text
URL: /
Component: home.index
Template: resource/templates/pages/home.gohtml
Props: HomeProps
JSON request: X-Goose-Page: true
Asset entry: resource/src/main.ts
```

### HomeProps

```go
type HomeProps struct {
    Sort         string
    Tabs         []TabPayload
    Topics       []TopicPayload
    Pagination   PaginationPayload
    Announcement AnnouncementPayload
}
```

首页 HTML 和 JSON payload 必须来自同一份 DTO。服务端 HTML 只在 `<noscript>` 中输出 no-js/SEO 所需的核心公开内容：页面标题、描述、主题链接、摘要、分类链接、基础统计和下一页链接。不要额外输出 hidden SEO 副本；完整布局、交互状态和视觉细节由 `resource/src` 的前端应用负责。

## 文章详情页面契约

```text
URL: /p/post/:id
Component: article.detail
Template: resource/templates/pages/article.gohtml
Props: ArticleDetailProps
JSON request: X-Goose-Page: true
Asset entry: resource/src/main.ts
```

### ArticleDetailProps

```go
type ArticleDetailProps struct {
    Article     ArticlePayload
    Replies     []ReplyPayload
    Permissions ArticlePermissions
}
```

详情页服务端 HTML 只在 `<noscript>` 中输出公开可读核心内容：标题、描述、作者链接、分类链接、渲染后的正文 HTML、回复文本和基础统计。回复编辑器、点赞、收藏、右侧栏、时间线等完整阅读交互由前端应用负责。

## 测试策略

Go 测试：

- V2 模板可以完整加载。
- 普通请求返回 shell HTML。
- `X-Goose-Page: true` 返回 JSON payload。
- 普通 HTML 中包含 `<noscript>` fallback，禁用 JavaScript 时公开页面仍有核心内容。
- 不存在公开 `?render=seo` 协议依赖。
- 分类 slug、权限、登录状态与旧逻辑一致。

前端测试：

- 初始化 payload 解析。
- 链接拦截规则。
- history 前进后退。
- title/meta/canonical 更新。
- 侧边栏 active 更新。
- 请求失败后原生跳转 fallback。

浏览器验收：

- 桌面和移动端截图。
- 禁用 JavaScript 后公开页面仍有核心内容。
- 文章详情 markdown、代码块、表格、Mermaid 等内容正确。
- 无明显首屏闪烁。

## 验收标准

迁移完成前，每个页面至少满足：

- 功能覆盖老页面。
- 公开内容有服务端 HTML。
- head meta、canonical、title 正确。
- 登录、权限、错误状态正确。
- JS 启用后站内跳转体验顺滑。
- JS 禁用后公开内容不空白。
- UI 符合 `RESOURCE_V2_UI_SPEC.md`。
- 不依赖旧 Alpine 页面 handler。
- 不在 Go 代码中出现样式选择器和前端布局参数。

## 删除旧实现标准

只有满足以下条件后才删除旧前台：

- 正式路由全部切到 V2。
- V2 页面和接口通过测试。
- 浏览器验收覆盖主要页面。
- 模板引用中没有旧前台依赖。
- 静态资源入口中没有旧页面脚本依赖。
- 文档中记录了替代关系。

删除时应一次性清理：

- 旧模板。
- 旧页面脚本。
- 不再使用的 Alpine handler。
- 不再使用的 CSS。
- 不再使用的 Go helper 或模板函数。

## 风险与处理

### SEO 内容漂移

风险：JS 页面和 no-js/SEO 页面内容不一致。

处理：共用服务端 DTO 和模板 partial，测试核心字段。

### 首屏闪烁

风险：服务端 HTML 和前端接管后的 DOM 差异过大。

处理：首屏 HTML 输出核心结构，前端增强时复用同一份 payload，不在加载后重排主要骨架。

### 新旧实现长期并存

风险：`resource` 与 `resource` 同时维护过久，形成两套系统。

处理：明确迁移批次和删除标准，完成后切换正式路由并删除旧实现。

### UI 继续发散

风险：每个页面临时写一套视觉处理。

处理：以 UI 规范为执行标准，页面实现前先补齐组件和状态规则。

## 实施阶段

### Phase 0：设计冻结

- 确认 `/v2` 临时前缀。
- 确认 `resource` 独立边界。
- 确认不公开 `?render=seo`。
- 确认 UI 规范。
- 当前阶段已完成，后续文档以已验收实现为准持续校正。

### Phase 1：基础设施

- 新增 V2 模板注册。
- 新增 V2 PagePayload。
- 新增 shell 渲染。
- 新增前端启动和站内导航基础能力。
- 新增 no-js/SEO 渲染测试。

### Phase 2：核心阅读链路

- 首页。
- 分类页。
- 文章详情页。
- 侧边栏。
- 文章 markdown/prose 体验。

### Phase 3：公开辅助页面

- 搜索。
- 用户主页。
- 友情链接。
- 赞助。
- 关于。

### Phase 4：登录态页面

- 通知。
- 设置。
- 私信。

### Phase 5：发布和编辑

- 发布页。
- 编辑页。
- markdown 预览。
- 草稿和校验。

### Phase 6：正式切换

- 正式路由切到 V2。
- 保留短期回退。
- 删除旧前台实现。
- 更新文档。

## 当前维护原则

V2 已进入实现验收和渐进补齐阶段。后续维护两份文档时，应以已验收的代码效果为准；如果文档与实现冲突，优先确认实现是否已验收，再修正文档：

- `RESOURCE_V2_DESIGN.md`：架构、SEO、迁移边界。
- `RESOURCE_V2_UI_SPEC.md`：视觉、交互、组件和页面规范。

新页面或新交互进入实现前，仍应先补齐规范；已经验收的页面效果不因旧文档描述而回退。

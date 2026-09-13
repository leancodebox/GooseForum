# React + shadcn 迁移与 Next 并存架构

## 状态

- 决策状态：方向已确定，隔离的 React workspace 与开发入口已经初始化。
- 下一步：将当前 Vue 客户端迁移为 React，并采用 React 官方 shadcn/ui 生态。
- 长期并存能力：Next 作为可选渲染方式接入，不取代 Go SSR，也不是 React 迁移的下一阶段或完成条件。
- 历史基线：`docs/shadcn-vue-site-migration.md` 记录了当前 shadcn-vue/Reka UI 基础层的迁移结果；该实现继续作为迁移前的行为与视觉基线。

## 背景

GooseForum 当前由 Go 提供页面 payload、HTML 模板、SEO/no-js 内容与业务 API，浏览器端由 Vue 3 应用接管主站和管理后台。基础 UI 已统一到 shadcn-vue 与 Reka UI，但 Vue 侧的 shadcn 生态在官方组件、registry、示例、第三方模板和工具支持方面弱于 React 侧。

迁移的直接目标是使用 React 官方 shadcn/ui 生态，同时保留现有 Go 页面渲染体系。Next 的目标不同：它是与 Go SSR 并存的另一种可选渲染方式，未来可由 GooseForum 主进程按配置启动并承接指定页面路由。

## 核心决策

1. React + shadcn 是当前前端演进方向。
2. Go 继续拥有业务逻辑、认证授权、数据库访问、API、文件服务和现有 SSR/payload 能力。
3. React 业务组件不得直接依赖 Next；通过小型运行时适配层接入不同宿主。
4. Go SSR 是长期受支持的运行方式，不视为等待 Next 替换的临时实现。
5. Next 是可选并存能力。未启用 Next 时，程序必须能够只依赖 Go 正常运行。
6. Next 不成为 React 迁移的验收条件；React 版本完成后即可独立交付。
7. 不在 Next 中复制 Go 业务逻辑。Next 只负责页面渲染、路由集成和必要的数据编排。

## 目标结构

```text
resource/
├── packages/
│   ├── client/             现有框架无关 payload、API 与浏览器协议
│   └── react/              React 业务组件、hooks、shadcn/ui 与宿主适配接口
├── apps/
│   ├── react/              C 端与后台双入口的 Vite 开发应用和 Go payload 适配
│   └── next/               可选 Next App Router 适配与服务端入口
├── templates/              长期保留的 Go HTML/SEO/no-js 模板
└── static/                 Go 模式使用的浏览器端构建产物
```

具体目录可以在实现阶段根据 pnpm workspace 调整，但依赖方向必须保持：

```text
@gooseforum/client
        ↑
@gooseforum/react
    ↑          ↑
Go 客户端适配   Next 适配
```

`@gooseforum/react` 不能反向导入 Next 应用代码。

## React 运行时适配层

适配层只解决宿主差异，不抽象论坛业务。候选边界包括：

- 页面 payload 的初始读取和后续获取。
- 链接渲染与客户端导航。
- 当前路径、查询参数和滚动恢复。
- locale、主题和页面元数据的宿主接入。
- 服务端与浏览器环境差异。

页面数据结构、权限判断结果和 API 请求继续使用 `@gooseforum/client`。主题 token 和现有视觉规范继续作为迁移基线。

避免为了同时支持两个宿主而构造庞大的通用运行时。普通组件应优先接收数据与回调；只有确实依赖路由或宿主环境的能力才进入适配接口。

## React + shadcn 迁移原则

1. 使用 React 官方 shadcn/ui 组件源码和官方支持的 primitives。
2. 不逐文件翻译当前 `resource/src/components/ui`。根据真实调用重新生成组件，再迁移 GooseForum 的主题、variant 和产品行为。
3. 保持现有页面信息架构、内容密度、键盘行为、焦点管理和移动端体验。
4. 复用框架无关代码，包括 `@gooseforum/client`、API、Markdown、ProseMirror、图片处理、格式化和 payload 类型。
5. React 业务组件不导入 `next/link`、`next/navigation`、Server Actions 或 Next 缓存 API。
6. Next 专属的 Server Component 边界、metadata 和路由文件只存在于 Next 适配应用。
7. 迁移采用纵向页面切片；完成一个切片后删除对应的 Vue 调用和无用样式。

## 实施顺序

### 阶段一：React 基础设施

- 建立 React workspace、TypeScript 和构建入口。
- 使用同一 Vite 应用的两个 HTML 入口：根 `index.html` 对应 C 端，`admin/index.html` 对应管理后台；两套源码分别位于 `src/site/` 和 `src/admin/`。
- 未检测到 Go 注入 payload 时，通过 Vite 代理读取真实 Go 页面 payload。
- React 开发构建产物保留在新 app 自己的 `dist/`，不进入现有 `resource/static/` 或 Go `embed.FS`。
- 初始化官方 shadcn/ui，映射现有主题 token；管理后台 shell 以官方 `dashboard-01` block 为起点，C 端按 GooseForum 的信息架构自行设计。
- 建立 Go payload、导航、i18n、主题和 flash message 的 React 适配。
- 保持现有 Go 模板与 `X-Goose-Page` 协议。
- 为核心交互增加浏览器级回归测试。

### 阶段二：页面迁移

建议按风险从低到高迁移：

1. 登录、重置密码、目录和静态信息页面。
2. 管理后台 CRUD 页面。
3. 首页、分类、搜索和用户页。
4. 设置、通知、消息、审核和发布。
5. Topic、PostComposer、编辑器、滚动窗口和内容增强。
6. AppShell 与共享状态收尾，移除 Vue 运行时。

该阶段完成后，Go SSR + React 即为独立、完整的目标运行方式，无需等待 Next。

### 独立工作流：Next 并存能力

Next 接入建立在可复用的 React 组件和适配边界之上，但属于独立能力，不是上述迁移的后续替换阶段。

预期运行拓扑：

```text
浏览器
  └── Go 公共监听端口
      ├── /api/*、/file/*、/oauth2/*、SEO 资源  → Go
      ├── Go 渲染路由                         → Go SSR + React
      └── 配置给 Next 的页面与 /_next/*       → Next 子进程
```

约束如下：

- Go 是主进程和统一公开入口。
- Next 仅监听 loopback 地址。
- Go 根据明确配置启动 Next，并通过反向代理路由请求。
- Next 服务端携带原始 Cookie 向 Go 获取 payload；Go 始终是认证和权限事实来源。
- 必须设计内部回源标识或内部 page-data endpoint，避免 Next 请求再次被代理回 Next。
- 初期按固定配置或路由白名单选择渲染器，不做随机灰度。
- Next 不可用时采用“启动失败”还是“降级到 Go SSR”，必须由配置明确决定，不能静默切换。

## Next 构建与 embed 边界

Next SSR 是需要 Node 运行时的服务，不能像普通 Vite 静态资源一样直接从 Go `embed.FS` 执行。

预期方案是：

1. 构建 Next standalone 产物。
2. 将 standalone、`.next/static` 和 `public` 作为发布产物；如要求单 Go 文件携带，可将其嵌入 Go 二进制。
3. 启动时把嵌入产物释放到临时目录。
4. Go 通过 `exec` 启动 Next 服务，并管理健康检查、退出和清理。
5. Node 运行时采用宿主依赖、发布包 sidecar 或平台专属嵌入方案，另行决策。

“Next 产物可被 Go 携带”与“Next 不需要独立运行时”是两件事。是否把 Node 可执行文件也嵌入最终二进制，需要结合发布体积、平台矩阵、签名和安全更新成本单独评估。

## 配置方向

最终配置名称在实现时确定，语义至少应覆盖：

```toml
[frontend]
renderer = "go"

[frontend.next]
enabled = false
command = "node"
host = "127.0.0.1"
startupTimeout = "15s"
failurePolicy = "fail-startup"
routes = []
```

其中：

- `renderer = "go"` 表示使用长期支持的 Go SSR + React 路径。
- 启用 Next 不代表关闭或删除 Go SSR。
- `routes` 允许稳定地把部分页面交给 Next；空列表的具体含义必须在实现时定义清楚。
- `failurePolicy` 必须显式配置是否允许回退。

## 非目标

- 不以 React 迁移为契机重写 Go 业务服务。
- 不要求把所有数据修改改造成 Next Server Actions。
- 不把 Next 作为 GooseForum 的唯一部署方式。
- 不在 React 迁移阶段实现 Next 进程管理和代理。
- 不承诺首版就把 Node 运行时嵌入单个 Go 二进制。
- 不同时长期维护两套 React 业务组件。

## 验收标准

### React + shadcn

- Go SSR、SEO/no-js 模板与现有 URL 行为保持有效。
- 主站和管理后台达到当前 Vue 版本的功能与无障碍基线。
- 明亮、暗色和自定义主题表现一致。
- 页面导航、返回滚动、登录态、编辑器和上传流程通过回归测试。
- Vue 依赖和对应实现可以从最终 React 构建中移除。
- Next 未实现或未启用时不影响 React 版本交付。

### Next 并存能力

- 配置关闭时不启动 Node/Next，Go 模式独立工作。
- 配置开启时由 Go 启动、检查和停止 Next 子进程。
- `/api`、文件、OAuth/OIDC 和认证仍由 Go 处理。
- Cookie、状态码、重定向、流式响应和代理头行为正确。
- Next 故障策略符合配置，且不会形成代理循环。
- 同一业务组件不因 Next 接入产生第二份实现。

## 待后续决策

- Next 路由按全局模式、固定白名单还是更细粒度规则切换。
- Node 由宿主提供、作为 sidecar 发布，还是按平台嵌入。
- Next 启动失败时的默认策略。
- Next 内部回源使用独立 loopback endpoint 还是受保护的内部请求头。
- Next 缓存、个性化 payload 和多实例部署策略。

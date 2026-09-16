# 前端结构

Go 负责业务、认证授权、页面 payload、HTML/SEO 和静态资源服务。浏览器应用通过 pnpm workspace 组装以下包。

| 目录 | 包名 | 职责 |
| --- | --- | --- |
| apps/web | @gooseforum/web | 主站与后台双入口、主站导航与缓存、后台页面、Go/Vite 浏览器宿主 |
| apps/next | @gooseforum/next | 可选的独立 Next.js App Router 宿主与 payload 适配器 |
| packages/client | @gooseforum/client | 框架无关的 API、payload 协议、类型与翻译数据 |
| packages/markdown | @gooseforum/markdown | 主站与后台共用的 Markdown 解析、编辑辅助与排版样式 |
| packages/runtime | @gooseforum/runtime | React Context/hooks、导航接口、页面预加载、语言运行时 |
| packages/ui | @gooseforum/ui | shadcn 基础组件、通用工具、基础样式和语义 token 默认值 |
| packages/theme-default | @gooseforum/theme-default | 默认论坛主题的页面、业务组件、排版与主题样式 |
| templates | — | Go HTML、SEO 和 no-js 内容 |
| static/dist/react | — | 生产构建产物，沿用 /assets/react/ URL |

依赖方向：web/next → theme-default → runtime/client/ui；runtime → client；ui 不依赖论坛主题或业务 API。Vite 后台和浏览器 host 不导入 theme-default；Next 只依赖共享包，不导入 Vite host。后续主题可替换 theme-default，继续复用相同的数据和运行时接口。当前不提供动态主题安装或插件注册机制。

`apps/web/src/host` 实现 Go 注入 payload、Vite 开发代理、浏览器 Cookie/主题状态及启动错误提示。`apps/web/src/site` 负责选用默认主题、导航、缓存与滚动；`apps/web/src/admin` 拥有独立后台。

UI 基础 CSS 不自动扫描整个仓库。默认主题和后台各自声明 Tailwind 来源并导入 UI 基础样式。主题专属样式不进入后台。主题包下的 test 包含页面及跨 runtime/UI 的集成回归；覆盖率继续统计拆分前的全部共享源码，避免目录迁移掩盖未覆盖代码。

从 resource 执行：

```sh
pnpm install --frozen-lockfile
pnpm dev             # client watch + web Vite，端口 3011
pnpm build           # 构建 client 与 web，再编译 Go 以嵌入产物
pnpm test            # client 与 React 测试
pnpm check:react:full
pnpm next:dev        # 独立 Next 服务，端口 3012
pnpm next:dev:go     # Next 连接本机 5234 端口的 GooseForum
pnpm check:next      # Next 类型、构建与浏览器验证
```

添加基础组件：`pnpm dlx shadcn@latest add <component> -c apps/web`。组件应落入 packages/ui；业务页面应留在对应主题或后台。

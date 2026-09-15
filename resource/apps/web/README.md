# GooseForum React 开发入口

该应用同时提供独立 Vite 开发宿主和 Go 模板生产宿主。开发时由 Vite 的 `index.html` 启动并通过代理读取 Go 页面数据；生产构建写入 `resource/static/dist/react/`，由 Go 模板通过 manifest 加载并注入首屏 payload。

## 启动

先保持 GooseForum Go 服务运行在默认的 `http://127.0.0.1:5234`，再执行：

```bash
cd resource
pnpm --filter @gooseforum/web dev
```

React 开发环境包含两个入口：

- C 端：`http://localhost:3011/`，HTML 为根目录 `index.html`，代码位于 `src/site/`。
- 管理后台：`http://localhost:3011/admin`，HTML 为 `admin/index.html`，代码位于 `src/admin/`。

独立 Vite 入口没有 `#goose-payload`，仅在开发模式通过 `/__goose_page/*` 代理读取真实 Go 页面数据。Go 模板宿主则必须注入 `#goose-payload`，生产包不会访问开发代理。C 端直接打开 `/categories` 等页面时，Vite 会返回根 `index.html`。后台是纯 SPA；开发服务器会将 `/admin` 和所有 `/admin/*` 页面导航固定回退到 `admin/index.html`，再由后台路由接管。`admin.shell` 只作为 viewer、权限、主题等启动数据，不表示后台采用页面 SSR。

Vite 宿主负责页面接口请求、History 和浏览器 metadata；`@gooseforum/theme-default` 只接收页面 payload 与导航/API runtime，不依赖 Vite、Go 模板或 Next。Go 模板与独立 Vite 页面共用浏览器宿主适配。

React 国际化使用 `i18next` 与 `react-i18next`。框架无关的 locale、标准化规则和翻译资源放在 `@gooseforum/client/i18n`；`@gooseforum/runtime/i18n` 为每个宿主创建独立 i18next 实例并提供 `I18nextProvider`。Vite 当前根据 query、Cookie 和浏览器语言选择初始 locale；未来 Next 应按请求创建实例，不能共享服务端全局实例。

如果 Go 服务使用其他地址：

```bash
GOOSE_DEV_ORIGIN=http://127.0.0.1:6234 pnpm --filter @gooseforum/web dev
```

## 验证

```bash
# 日常门禁：类型、ESLint/React Compiler、单测、DOM、死代码和包体积
pnpm check:react

# 完整门禁：在日常门禁上增加覆盖率和 Chromium 桌面/移动端烟测、axe WCAG 检查
pnpm check:react:full

# 全工作区生产依赖漏洞
pnpm audit:dependencies
```

ESLint 会以当前迁移存量为告警上限，新增告警会失败；Vitest 对 app 与共享 React 包分别设置覆盖率下限。Playwright 测试使用拦截的页面 payload，不依赖本地 Go 服务或测试数据库。构建检查同时限制首屏依赖总量和最大的懒加载 chunk。

React 构建产物写入 `resource/static/dist/react/`，生产 Go 二进制会嵌入对应 manifest 与资源；前端构建后需要重新构建 Go 二进制。

## 添加 shadcn/ui 组件

从 `resource` 目录执行：

```bash
pnpm dlx shadcn@latest add <component> -c apps/web
```

基础 UI 源码会根据 `components.json` 写入 `resource/packages/ui/src/components/`。应用级 block 默认写入 `src/admin/components/`；C 端产品组件由 GooseForum 自行设计并放入 `src/site/`。

管理后台的初始布局来自官方 `dashboard-01`：

```bash
pnpm dlx shadcn@latest add dashboard-01 -c apps/web
```

后台业务页面通过 GooseForum payload 与 API 提供数据。

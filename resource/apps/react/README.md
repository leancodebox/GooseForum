# GooseForum React 开发入口

该应用是独立于现有 Vue 入口的 React + Vite 开发环境。它不会被当前 Go 模板、生产资源构建或 `go:embed` 自动加载。

## 启动

先保持 GooseForum Go 服务运行在默认的 `http://127.0.0.1:5234`，再执行：

```bash
cd resource
pnpm --filter @gooseforum/react-app dev
```

React 开发环境包含两个入口：

- C 端：`http://localhost:3011/`，HTML 为根目录 `index.html`，代码位于 `src/site/`。
- 管理后台：`http://localhost:3011/admin/`，HTML 为 `admin/index.html`，代码位于 `src/admin/`。

两个入口都通过 `/__goose_page/*` 代理读取对应的真实 Go 页面 payload。C 端直接打开 `/categories` 等页面时，Vite 会返回根 `index.html`；后台开发入口使用带尾斜杠的 `/admin/`。

如果 Go 服务使用其他地址：

```bash
GOOSE_DEV_ORIGIN=http://127.0.0.1:6234 pnpm --filter @gooseforum/react-app dev
```

## 验证

```bash
pnpm --filter @gooseforum/react typecheck
pnpm --filter @gooseforum/react-app typecheck
pnpm --filter @gooseforum/react-app build
```

React 构建产物写入本目录的 `dist/`，不会进入当前 Go 内嵌的 `resource/static/`。

## 添加 shadcn/ui 组件

从 `resource` 目录执行：

```bash
pnpm dlx shadcn@latest add <component> -c apps/react
```

基础 UI 源码会根据 `components.json` 写入 `resource/packages/react/src/components/ui/`。应用级 block 默认写入 `src/admin/components/`；C 端产品组件由 GooseForum 自行设计并放入 `src/site/`。

管理后台的初始布局来自官方 `dashboard-01`：

```bash
pnpm dlx shadcn@latest add dashboard-01 -c apps/react
```

当前保留了 block 的示例数据和内容，用于确认布局、响应式侧栏、图表和表格链路。迁移具体后台页面时再逐项替换为 GooseForum payload 与 API。

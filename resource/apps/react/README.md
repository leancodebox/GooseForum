# GooseForum React 开发入口

该应用是独立于现有 Vue 入口的 React + Vite 开发环境。它不会被当前 Go 模板、生产资源构建或 `go:embed` 自动加载。

## 启动

先保持 GooseForum Go 服务运行在默认的 `http://127.0.0.1:5234`，再执行：

```bash
cd resource
pnpm --filter @gooseforum/react-app dev
```

React 开发入口位于 `http://localhost:3011`。直接打开 `/categories` 等页面时，Vite 会返回本应用的 `index.html`，随后通过 `/__goose_page/*` 代理读取对应的真实 Go 页面 payload。

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

基础 UI 源码会根据 `components.json` 写入 `resource/packages/react/src/components/ui/`。

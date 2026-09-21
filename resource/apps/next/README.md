# GooseForum Next host

独立的 Next.js App Router 宿主，用于验证默认主题和 `@gooseforum/runtime` 可以脱离 Vite/Go 模板运行。

```sh
cd resource
pnpm next:dev       # http://127.0.0.1:3012
pnpm next:dev:go    # 连接本机 5234 端口的 GooseForum
pnpm check:next     # 类型检查、production build 与浏览器验证
pnpm next:start --goose-origin http://127.0.0.1:5234
```

默认不依赖 Go，使用内置只读 demo payload。设置 `GOOSEFORUM_ORIGIN` 后，Next 服务端会携带请求 Cookie 和语言头，通过 `X-Goose-Page` 协议读取真实页面 payload：

```sh
GOOSEFORUM_ORIGIN=http://127.0.0.1:5234 pnpm next:dev
```

连接 upstream 时，Next 会把 `/api`、文件、OAuth、静态资源和主题 CSS 透明代理到同一地址。首屏由 Next SSR。独立启动时，水合后的站内导航通过 `/goose-page-data` 获取 JSON payload；由 Go 管理启动时，浏览器直接携带 `X-Goose-Page` 请求 Go，不再经过 Next。

在 `config.toml` 的 `[server]` 中设置 `next = true` 后，Go 主进程会启动和停止
standalone 子进程，并将主站页面转发给 Next；Node 环境或构建产物不可用时自动回退到
Go 前端。发布构建使用 `go generate ./...` 生成并嵌入 standalone ZIP；运行时释放到
独立临时目录，进程退出后自动清理。本地没有内嵌 ZIP 时会使用工作区 `.next/standalone`。

部署 standalone 目录后可直接运行：

```sh
node start.mjs --goose-origin http://127.0.0.1:5234 --hostname 0.0.0.0 --port 3012
```

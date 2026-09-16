# GooseForum Next host

独立的 Next.js App Router 宿主，用于验证默认主题和 `@gooseforum/runtime` 可以脱离 Vite/Go 模板运行。

```sh
cd resource
pnpm next:dev       # http://127.0.0.1:3012
pnpm next:dev:go    # 连接本机 5234 端口的 GooseForum
pnpm check:next     # 类型检查、production build 与浏览器验证
pnpm next:start -- --goose-origin http://127.0.0.1:5234
```

默认不依赖 Go，使用内置只读 demo payload。设置 `GOOSEFORUM_ORIGIN` 后，Next 服务端会携带请求 Cookie 和语言头，通过 `X-Goose-Page` 协议读取真实页面 payload：

```sh
GOOSEFORUM_ORIGIN=http://127.0.0.1:5234 pnpm next:dev
```

连接 upstream 时，Next 会把 `/api`、文件、OAuth、静态资源和主题 CSS 透明代理到同一地址。首屏由 Next SSR；水合后的站内导航继续通过 `/goose-page-data` 获取 JSON payload，不再经过 Next 页面渲染。

当前应用不包含 Go 子进程管理或公开入口的渲染路由切换；这些属于后续集成层。

部署 standalone 目录后可直接运行：

```sh
node start.mjs --goose-origin http://127.0.0.1:5234 --hostname 0.0.0.0 --port 3012
```

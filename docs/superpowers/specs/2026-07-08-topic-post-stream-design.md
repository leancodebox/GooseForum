# Topic/Post Stream 接口设计

## 背景

GooseForum 已经逐步迁移到 topic/post 模型：

- topic 是讨论容器，负责标题、分类、统计、状态等元信息。
- 第一个 post（`post_no = 1`）是主题正文。
- 后续 post（`post_no > 1`）是普通回复。

但当前详情页流程仍然保留了一些 article/reply 时代的结构：

- `TopicDetail` 单独加载首楼 post，并把渲染后的正文暴露为 `topic.html`。
- `PostWindow` 明确排除 `post_no <= 1`。
- 前端把主题正文和回复列表分开渲染，再通过 `data-post-no="1"` 补齐时间轴行为。

这套实现可以工作，但服务端接口层和数据模型之间还不够一致。

## 参考模型

Discourse 把主题正文视为 topic 的第一个 post：

- topic 元信息和 post stream 是两个独立概念。
- topic 详情响应包含 topic 元信息和初始 post stream。
- 首楼是一个普通 post，只在需要时拥有额外的 topic 级行为。
- 分页、楼层跳转、删除后的楼层空洞、时间轴，都围绕 post stream 运转。

GooseForum 应该参考这个结构，但不需要照搬 Discourse 的全部复杂能力。

## 目标

围绕统一的 post stream 模型合并 topic 详情和 post window 的服务端行为。

目标服务端形态是：

```text
TopicDetail = Topic 元信息 + 初始 PostWindow
PostWindow  = 同一条 post stream 的后续窗口
```

首楼必须作为 `postNo = 1` 的 `PostPayload` 返回，而不是继续放在 `topic.html` 里。

## 非目标

- 本阶段不修改公开路由。
- 本阶段不引入 Discourse 那种复杂的任意 post ID 批量 stream API。
- 本阶段不修改存储结构；`topics` 和 `posts` 已经是目标模型。
- 本阶段不额外增加路由兼容层。

## 目标 Payload

`TopicDetailProps` 应调整为：

```ts
interface TopicDetailProps {
  topic: TopicDetailPayload
  postStream: PostWindowPayload
  hotTopics: TopicPayload[]
  permissions: PagePermissions
}
```

`TopicDetailPayload` 只保留 topic 元信息，不再暴露正文 HTML。

`PostWindowPayload` 作为可复用的 stream 响应：

```ts
interface PostWindowPayload {
  posts: PostPayload[]
  anchorPostId: number
  beforeCursor: number
  afterCursor: number
  beforePostNo: number
  afterPostNo: number
  hasBefore: boolean
  hasAfter: boolean
  total: number
  maxPostNo: number
}
```

`posts` 可以包含 `postNo = 1`。如果某个调用方只需要回复，应通过请求 `postNo > 1` 的窗口或在查询层过滤，而不是依赖 payload 构造函数隐藏首楼。

## 服务端设计

### Post Stream 构造器

抽出一个共享的 post stream 构造器，同时服务于 `TopicDetail` 和 `PostWindow`。

职责：

- 加载某个 topic 的 post 窗口。
- 保留真实 `post_no`。
- 支持 `post_no = 1`。
- 按请求模式处理删除后的楼层空洞，例如向前或向后寻找可用 post。
- 统一构造 `PostPayload`。
- 返回游标和时间轴元数据。

现有 `PostWindow` 控制器应变成这个构造器的轻量封装。

`TopicDetail` 则负责加载 topic 元信息，然后用同一个构造器获取初始 post stream。

### Payload 构造

`buildPostPayloads` 不应再硬编码排除 `post_no <= 1`。

过滤应该在查询或窗口语义层显式完成：

- topic 详情初始窗口包含首楼。
- post window 请求根据游标语义决定是否包含首楼。
- 如果仍有 reply-only 场景，则应查询 `post_no > 1`。

### 首楼特殊行为

首楼在内容渲染、时间轴、举报、正文编辑方面都是普通 post。

但它仍然需要 topic 级行为：

- 标题、分类、标签展示。
- topic 级编辑操作。
- topic 级审核操作。
- topic 级分享、SEO、meta。

这些特殊行为应通过前端的 `post.postNo === 1` 判断和服务端提供的 topic 权限共同处理。

## 前端设计

`TopicPage` 应渲染一条统一的 post stream：

- 从 `page.props.postStream.posts` 初始化。
- `postNo = 1` 渲染为主题正文。
- `postNo > 1` 渲染为普通回复。
- 时间轴和跳楼逻辑继续基于真实 `postNo`。

单独渲染 `topic.html` 的正文块应被移除。

首楼可以有不同的按钮和标题区域，但它仍然应该处在同一个列表中，参与同一套滚动、定位和时间轴逻辑。

## 兼容性

这是站点前端和服务端之间的内部契约调整。前后端会一起发布，因此可以干净地更新接口。

为了降低实现风险，可以短期保留临时兼容字段；但最终目标应该移除 `topic.html` 等旧字段。

## 测试

后端测试应覆盖：

- topic 详情返回的 post stream 第一项是 `postNo = 1`。
- post window 可以锚定到 `postNo = 1`。
- post window 在删除楼层造成空洞时仍能正确向前或向后寻找。
- 游标和 `maxPostNo` 使用真实 post number。
- 首楼正文通过 `PostPayload` 渲染返回。

前端测试和构建检查应覆盖：

- topic 详情页从同一条 stream 渲染首楼和回复。
- 后续加载回复不会重复插入首楼。
- 时间轴从 1 开始，并且删除回复后仍保持真实楼层号。
- 编辑、举报、审核等控件仍能区分首楼和普通回复。

## 实施顺序

1. 在服务端新增或抽出共享 post stream 构造器。
2. 将 `TopicDetailProps` 调整为暴露 `postStream`。
3. 允许 `PostPayload` 构造逻辑处理 `postNo = 1`。
4. 将 `PostWindow` 改为复用同一套 stream 逻辑。
5. 更新前端类型和 `TopicPage`，改为从统一 stream 渲染。
6. 前端迁移完成后移除 `topic.html` 使用。
7. 运行后端测试、前端测试、前端构建和 diff 检查。


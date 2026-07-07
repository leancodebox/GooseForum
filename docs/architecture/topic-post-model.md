# Topic / Post Model Refactor Draft

本文档整理将现有 `articles` / `reply` 模型改造为 `topics` / `posts` 模型的设计思路。它是评估和后续拆分实施的基础，不代表立即执行完整迁移。

## 背景

当前 GooseForum 的正文数据分散在两个模型中：

- `articles` 同时承担主题容器和首楼正文。
- `reply` 承担后续回复正文。

这导致首楼和回复在很多能力上需要分别处理，例如渲染、编辑、审核、举报、搜索、文件引用、通知、用户动态和统计。随着回复编辑、举报、文件引用等能力逐渐增强，这种分叉会继续放大维护成本。

目标模型是：

- `topic` 表示一个讨论主题，是列表、分类、状态和统计的容器。
- `post` 表示一条正文，首楼和普通回复都属于 `post`。
- 一个 `topic` 至少有一个首楼 `post`。

## 核心原则

1. **正文只属于 post**

   `topic` 不保存正文的权威数据。`topic` 可以缓存摘要、首图、最后回复时间等列表展示字段，但这些字段都应来自 `post` 或由 `post` 派生。

2. **首楼也是 post**

   首楼使用 `post_no = 1`。普通回复从 `post_no = 2` 开始。这样编辑、渲染、审核、举报、文件引用都可以统一到 post 维度。

3. **topic 负责讨论容器状态**

   topic 负责标题、分类、发布状态、置顶、浏览量、回复数、最后活跃时间等容器级信息。

4. **先兼容，再替换**

   这是核心数据模型改造，不应一次性删除旧表。应通过兼容层和分阶段迁移降低风险。

## 建议模型

### topics

```text
topics
  id
  title
  type
  user_id
  category_ids
  status
  process_status
  pin_weight
  view_count
  post_count
  reply_count
  first_post_id
  last_post_id
  last_posted_at
  excerpt
  first_image_url
  posters
  created_at
  updated_at
  deleted_at
```

字段说明：

- `title`：主题标题。
- `type`：主题类型，对应现有文章类型。
- `user_id`：主题创建者，通常等于首楼作者。
- `category_ids`：主题分类。现有 `article_category_rs` 后续可迁为 `topic_category_rs`。
- `status`：草稿、发布等发布状态。对应现有 `article_status`。
- `process_status`：管理状态，例如正常、封禁。
- `post_count`：主题内所有可计数 post 数，包含首楼。
- `reply_count`：回复数，可等于 `post_count - 1`，为了列表兼容可以保留缓存字段。
- `first_post_id`：首楼 post。
- `last_post_id`：最后活跃 post。
- `last_posted_at`：最后发帖时间，用于列表排序。
- `excerpt` / `first_image_url`：列表、SEO、分享用缓存字段，从首楼 post 派生。
- `posters`：参与者缓存，延续现有列表展示能力。

### posts

```text
posts
  id
  topic_id
  post_no
  user_id
  content
  rendered_html
  rendered_version
  reply_to_post_id
  process_status
  created_at
  updated_at
  deleted_at
```

字段说明：

- `topic_id`：所属 topic。
- `post_no`：主题内稳定楼层号。首楼为 1。
- `user_id`：post 作者。
- `content`：Markdown 原文。
- `rendered_html` / `rendered_version`：渲染缓存。
- `reply_to_post_id`：回复引用的上级 post，可替代现有 `reply.reply_id`。
- `process_status`：post 级管理状态，例如正常、封禁。

## 现有字段映射

### articles -> topics

| 现字段 | 目标字段 | 说明 |
| --- | --- | --- |
| `articles.id` | `topics.id` 或迁移映射 id | 可考虑保留 topic id 等于旧 article id，降低 URL 迁移成本。 |
| `title` | `topics.title` | 主题标题。 |
| `type` | `topics.type` | 主题类型。 |
| `category_id` | `topics.category_ids` | 后续可以拆成关系表。 |
| `user_id` | `topics.user_id` | 主题作者。 |
| `article_status` | `topics.status` | 建议重命名为更中性的 status。 |
| `process_status` | `topics.process_status` | 容器级管理状态。 |
| `view_count` | `topics.view_count` | 浏览量保留在 topic。 |
| `reply_count` | `topics.reply_count` | 兼容列表展示。 |
| `reply_seq` | `topics.post_seq` 或由 posts 计算 | 如果保留序列，应改名为 post_seq。 |
| `pin_weight` | `topics.pin_weight` | 全站置顶权重。 |
| `posters` | `topics.posters` | 参与者缓存。 |
| `description` | `topics.excerpt` | 从首楼 post 派生，也可继续缓存。 |
| `first_image_url` | `topics.first_image_url` | 从首楼 post 派生。 |
| `created_at` / `updated_at` | `topics.created_at` / `updated_at` | 容器时间。 |

### articles -> posts

| 现字段 | 目标字段 | 说明 |
| --- | --- | --- |
| `articles.id` | `posts.topic_id` | 首楼 post 归属的 topic。 |
| `articles.user_id` | `posts.user_id` | 首楼作者。 |
| `articles.content` | `posts.content` | 首楼正文。 |
| `articles.rendered_html` | `posts.rendered_html` | 首楼渲染缓存。 |
| `articles.rendered_version` | `posts.rendered_version` | 首楼渲染版本。 |
| `articles.process_status` | `posts.process_status` | 是否同步需要谨慎。topic 封禁和首楼封禁语义不完全等价。 |
| `articles.created_at` | `posts.created_at` | 首楼发布时间。 |
| `articles.updated_at` | `posts.updated_at` | 首楼更新时间。 |

首楼建议生成：

```text
posts.topic_id = articles.id
posts.post_no = 1
```

### reply -> posts

| 现字段 | 目标字段 | 说明 |
| --- | --- | --- |
| `reply.id` | `posts.id` 或迁移映射 id | 可保留旧 id 到新 post id 的映射表。 |
| `reply.article_id` | `posts.topic_id` | 所属主题。 |
| `reply.reply_no` | `posts.post_no` | 如果首楼为 1，旧 reply_no 可能需要整体 +1。 |
| `reply.user_id` | `posts.user_id` | 作者。 |
| `reply.content` | `posts.content` | 正文。 |
| `reply.rendered_html` | `posts.rendered_html` | 渲染缓存。 |
| `reply.rendered_version` | `posts.rendered_version` | 渲染版本。 |
| `reply.reply_id` | `posts.reply_to_post_id` | 引用的父回复。迁移时需要 id 映射。 |
| `reply.process_status` | `posts.process_status` | post 级管理状态。 |
| `reply.created_at` / `updated_at` | `posts.created_at` / `updated_at` | 时间。 |

## 路由边界

本轮模型评估先剔除路由改造。

现有公开路由和锚点保持不变：

```text
/p/post/{articleID}
/p/post/{articleID}#post-{postID}
```

即使内部逐步引入 `topic` / `post` 概念，也不在本轮设计中引入新的公开路由，例如 `/p/topic/{topicID}`。当前公开详情页 URL 继续保留 `/p/post/{id}`，但页面锚点已经切换到 `post-{id}`。

这样做的原因：

- 路由变化会额外牵涉 SEO、站内链接、通知链接、用户动态链接、外部已分享链接和浏览器锚点定位。
- 数据模型迁移本身已经足够大，路由改造会放大回滚成本。
- 现有 `/p/post/{id}` 虽然命名不够准确，但可以继续作为兼容入口。

因此后续讨论默认：

- `articleID` 在公开接口和 URL 上可以继续存在一段时间。
- 内部代码可以逐步把它解释为 `topicID`。
- 页面锚点统一使用 `#post-{postID}`。
- 是否改公开路由另起设计，不纳入本次改造范围。

## 影响面梳理

这次改造会影响面很广。下面按模块列出主要依赖点、迁移动作和风险。

| 模块 | 当前依赖 | 改造关注点 | 风险 |
| --- | --- | --- | --- |
| 文章列表 | `articles` 表、`article_status`、`process_status`、`reply_count`、`view_count`、`pin_weight`、分类关系 | 列表主体迁到 `topics`；摘要、首图、参与者继续作为 topic 缓存字段 | 列表排序、分类筛选、置顶、热门/最新查询的索引需要重新评估 |
| 详情首屏 | `articles.content`、`articles.rendered_html`、`reply.GetFirstPageByArticleId` | 首楼正文改为读取 `posts.post_no = 1`；回复读取同 topic 下其他 posts | 首楼和回复混排后，首屏 payload 字段需要兼容旧前端 |
| 回复分页 | `reply.article_id`、`reply.reply_no`、`reply.id` 游标 | 改为 `posts.topic_id`、`post_no`、`post.id`；仍兼容 `replyID` 锚点 | 游标方向、锚点定位、楼层号迁移最容易产生错位 |
| 发布主题 | 写 `articles`，正文直接落在 article | 创建 topic 后创建首楼 post；topic 缓存摘要、首图 | 草稿状态下是否创建 post、首楼编辑和 topic 标题编辑要拆清 |
| 回复发布 | 写 `reply`，更新 article reply 统计 | 创建普通 post，更新 topic reply/post 统计和参与者缓存 | 统计同步、通知、用户动态、文件引用需要跟随切换 |
| 编辑正文 | 文章编辑和回复编辑分两条逻辑 | 首楼和回复统一为 post 编辑；topic 标题/分类仍单独编辑 | 权限判断需要区分 topic 管理和 post 作者编辑 |
| 删除/软删 | `articles.DeletedAt`、`reply.DeletedAt` | topic 删除和 post 删除语义分离；删除首楼是否删除整个 topic 需要定义 | 删除首楼、删除最后回复、恢复数据都会影响统计 |
| 审核/封禁 | `articles.process_status`、`reply.process_status`、`moderationLog.SubjectArticle/Reply` | 正文级封禁迁到 post；topic 级处理保留在 topic | “封禁主题”和“封禁首楼”语义不同，不能简单合并 |
| 举报 | 旧 `article/reply` target 由迁移脚本读取，`article_id` 作为历史存储字段保留 | 正文举报迁到 `target_type=post`，主题举报使用 `target_type=topic`；保留 topic_id 便于范围查询 | 版主按分类过滤举报依赖 topic/category，新代码不再暴露旧 target 常量 |
| 文件引用 | 旧 `article/reply` target 由迁移脚本读取 | 正文内联图片迁到 `TargetTopic/TargetPost`；头像、管理上传不变 | 旧 target 只保留在迁移脚本中，新写入不再暴露旧常量 |
| 搜索 | 文章正文和标题搜索依赖 article；回复搜索目前较弱 | topic 索引标题/摘要；post 索引正文 | 搜索结果需要决定返回 topic 还是具体 post |
| 通知/未读 | `eventNotification` payload 使用 `topicId/postId/topicTitle`；链接拼 `#post-{id}` | 运行时代码只保留 topic/post 命名，历史数据通过迁移脚本转换 | 历史通知的存量数据需要通过迁移清洗完成 |
| 用户动态 | `userActivities.SubjectTopic`、`SubjectPost`，评论动态通过 reply 回查 article | SubjectPost 应指向 post；仍需能从 post 找 topic | 旧动态的 subject_id 是 reply id，迁移期必须有映射或兼容查询 |
| 用户统计 | `user_statistics.article_count/reply_count`、`articles_user_stat` | article_count 可变为 topic_count；reply_count 可变为 post_count 排除首楼 | 统计口径变化会影响个人页展示和徽章/积分逻辑 |
| 主题参与者 | `articles.posters` 由 `articles_user_stat` 同步 | topic 继续缓存 posters；来源改为 posts 聚合 | 删除/封禁 post 后参与者缓存是否重算要定义 |
| 分类 | `article_category_rs.article_id` | 可维持一段时间，内部解释为 topic_id；后续再改名为 topic_category_rs | 改表名收益不大，可放到最后做 |
| 点赞/收藏/关注 | `articleUserAction.article_id` | 这些是容器级动作，建议继续挂 topic | 命名可先不改表，避免牵连列表和用户页 |
| 浏览量 | `topicviewservice` 更新 topics.view_count | 浏览量属于 topic，迁到 topics.view_count | 热门排序依赖该字段，迁移需保持原子更新语义 |
| SEO / Sitemap / RSS | 使用 article title/content/description/first_image_url | title 属于 topic，正文摘要来自首楼 post，URL 不变 | 历史 URL 不变时 SEO 风险较低，但摘要生成路径要改 |
| HTTP 通知 | payload 使用 `topic` / `post` 字段，`targetType` 使用 `topic/post` | 新接入方只使用新字段，不再扩散旧命名 | 第三方回调消费者升级时需要同步修改字段读取 |
| 管理后台 | 文章管理、举报管理、审核日志读取 article/reply | 管理文章列表仍以 topic 为主，正文处理以 post 为主 | 文案和操作语义要避免“封禁主题”和“封禁回复”混淆 |
| 数据迁移 | 当前 `reply_sequence`、article user action 等迁移依赖 article/reply | 新增 posts 回填和 id 映射迁移；旧迁移保留 | SQLite/MySQL 差异、批量回填时间、失败重试需要单独设计 |
| 缓存 | 文章列表缓存、用户缓存、未读缓存等间接依赖 article/reply | 缓存 key 和失效点需要随着 topic/post 写路径调整 | 双写阶段最容易出现旧表新表缓存不一致 |

高风险点：

- `reply.reply_no` 到 `posts.post_no` 的迁移口径。
- 旧 `reply.id` 与新 `post.id` 的映射关系。
- 审核、举报、文件引用里的 `target_type` 迁移。
- 用户动态和通知里的历史链接兼容。
- 列表排序和回复分页的索引设计。
- 双写期间的统计和缓存一致性。

## 分阶段路线

### 阶段 0：命名适配，不改数据库

目标是降低后续认知成本。

- 在 service / payload 层引入 `Topic` / `Post` 语义。
- 新代码尽量使用 `topicID`、`postID`、`postNo`。
- 保持底层 `articles` / `reply` 表不变。
- 对外接口仍兼容现有字段。

这一阶段风险最低，可以作为准备工作。

### 阶段 1：新增 posts 表，首楼和回复双写或回填

目标是把正文能力统一到 post 维度。

- 新增 `posts` 表。
- 历史数据回填：
  - 每条 article 生成一个首楼 post。
  - 每条 reply 生成一个普通 post。
- 新增 id 映射能力：
  - article id -> topic id。
  - reply id -> post id。
- 新发布/回复可以先双写旧表和新表，或者继续写旧表后异步同步 posts。

这一阶段不建议马上切读路径。

### 阶段 2：详情页和回复列表读 posts

目标是让核心读路径切到新模型。

- 详情页首楼读取 `posts.post_no = 1`。
- 回复列表读取 `posts.topic_id = ? AND post_no > 1`。
- 楼层定位改为 post_no。
- 编辑首楼和编辑回复统一使用 post 更新逻辑。
- 渲染缓存统一基于 posts。

旧表仍保留，作为回滚和兼容基础。

### 阶段 3：周边能力迁移到 post

目标是消除 article/reply 分叉。

- 举报 target 从 `article/reply` 迁到 `post`，topic 级举报另行定义。
- 审核日志 subject 从 `article/reply` 迁到 `topic/post`。
- 文件引用 target 支持 `post`。
- 搜索索引以 topic + post 组合建立。
- 用户动态中的回复记录迁到 post。
- 通知中的 comment/reply 统一到 post。

这一阶段应按模块逐个迁移，不建议一次性完成。

### 阶段 4：清理旧表和旧字段

目标是最终收敛。

- `articles.content`、`articles.rendered_html`、`articles.rendered_version` 标记废弃或移除。
- `reply` 表停止写入，最终废弃。
- 旧 URL 和旧 id 映射保留一段时间。
- 删除兼容代码前需要确认线上无旧任务、旧通知、旧索引引用。

## 索引建议

posts 核心索引：

```text
idx_posts_topic_no(topic_id, post_no)
idx_posts_topic_id(topic_id, id)
idx_posts_topic_process(topic_id, process_status, post_no)
idx_posts_user_created(user_id, created_at)
idx_posts_reply_to(reply_to_post_id)
```

topics 核心索引：

```text
idx_topics_status_updated(status, process_status, pin_weight, updated_at, id)
idx_topics_status_hot(status, process_status, reply_count, id)
idx_topics_status_popular(status, process_status, view_count, id)
idx_topics_user_status(user_id, status, process_status, id)
```

具体索引需要结合 SQLite 和 MySQL 的查询计划再收敛，避免为了假想查询堆过多索引。

## 关键设计问题

### topic 的 process_status 和首楼 post 的 process_status 是否同步？

不建议完全同步。

- 封禁 topic：表示整个主题不可见或被处理。
- 封禁首楼 post：表示首楼内容不可见，但 topic 是否继续存在需要单独决定。

早期可以保持当前行为：封禁“文章”时同时影响 topic 展示和首楼 post 内容。长期应拆清语义。

### reply_no 是否要变成 post_no？

建议是。

如果首楼是 `post_no = 1`，现有 `reply.reply_no` 可以迁移为 `post_no = reply_no + 1`。这样前端显示楼层时更统一。

如果为了兼容现有展示，也可以保留：

- 首楼不显示楼层。
- 回复显示旧 reply_no。

但长期模型上建议统一为 post_no。

### 是否保留 reply_count？

建议 topic 保留 `reply_count` 缓存字段，因为列表排序和展示会频繁使用。

同时可以保留 `post_count`，用于更准确表达主题内 post 总数。两者关系通常是：

```text
post_count = reply_count + 1
```

### target_type 如何迁移？

不建议一次性把所有 target 都改成 post。

建议分层：

- 容器级能力：`topic`
- 正文级能力：`post`

例如：

- 收藏、关注、浏览：topic。
- 举报正文、封禁正文、文件引用：post。
- 分类、置顶、标题编辑：topic。

## 推荐下一步

短期不要直接迁移数据库。建议先做两件事：

1. 在代码层引入 `topic/post` 命名适配，减少新代码继续扩大 `article/reply` 心智负担。
2. 单独写一份迁移清单，逐项列出当前所有依赖 `article_id`、`reply_id`、`target_type=article/reply` 的模块。

等清单明确后，再决定是否进入新增 `topics/posts` 表的阶段。

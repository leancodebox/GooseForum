# Topic/Post 收尾操作文档

本文档记录当前 `article/reply` 到 `topic/post` 改造后的剩余任务。当前阶段的目标不是继续改功能，而是把已经完成的模型和 API 迁移做干净，降低后续维护时的命名混乱。

## 当前状态

已经完成：

- 新表和新模型：`topics`、`posts`、`category`、`topic_category_index`、`topic_user_action`、`topic_user_stat` 已经进入主要读写路径。
- C 端写操作 API 已改为 `topics/*` 和 `posts/*`。
- Admin 内容管理 API 已改为 `admin/topics/*`。
- 回复窗口和审核 API 已改为 `posts/window`、`moderation/topic-status`、`moderation/post-status`。
- 举报新写入已使用 `targetType = topic/post`。
- 历史举报、日志、通知会由迁移脚本转换到 `topic/post`；运行时代码后续不再保留 `article/reply` 兼容读取。
- 旧 model 目录暂时保留，不在当前阶段删除。

暂不处理：

- 路由页面路径 `/p/post/:id`。
- 旧数据库表的物理删除。
- 历史日志、历史通知、历史举报 payload 的强制迁移。

## 总体原则

- 每一轮只做一种类型的清理：前端命名、后端命名、历史兼容、旧 model 删除不要混在一个提交里。
- 不做机械全局替换。除了迁移期旧 model，运行时代码不再继续保留 `article/reply` 兼容字段或别名。
- 外部 API 已经切到 topic/post，后续不再新增 article/reply 语义的 API。
- 旧 model 目录最后再处理，只有确认运行时不依赖后再删。
- 每一阶段完成后都运行：

```bash
go test ./...
pnpm -C resource exec vitest run
pnpm -C resource build
```

`pnpm -C resource build` 目前会出现 `@vueuse/core` 的 Rolldown `INVALID_ANNOTATION` warning，但构建退出码为 0 时可接受。

## 阶段 1：前端详情页命名收敛

目标：前端代码内部尽量使用 `topic/post`，减少新代码继续沿用 `article/reply`。

建议修改范围：

- `resource/src/runtime/api.ts`
- `resource/src/types/payload.ts`
- `resource/src/site/pages/TopicPage.vue`
- `resource/src/site/pages/PublishPage.vue`
- 相关 site/admin 组件和类型。

建议操作：

1. 页面组件文件名已从 `ArticlePage.vue` 调整为 `TopicPage.vue`，详情页内嵌组件已调整为 `PostComposer.vue` / `PostPositionRail.vue`。
2. 把 runtime 中旧包装函数替换为：
   - `postReply` -> `createPost`
   - `updateReply` -> `updatePost`
   - `deleteReply` -> `deletePost`
   - `updateModerationArticleStatus` -> `updateModerationTopicStatus`
   - `updateModerationReplyStatus` -> `updateModerationPostStatus`
3. 页面详情 payload 使用 `topic/posts`，post 字段使用 `topicId/postNo`：
   - `replyId` -> `postId`
   - `replyNo` -> `postNo`，如果 UI 文案仍叫“回复楼层”，可只改代码变量。
   - `currentArticleId` -> `currentTopicId`
   - `articleProcessStatus` -> `topicProcessStatus`
4. 类型使用新名字，不再额外保留详情页兼容别名：
   - `ArticlePayload` -> `TopicDetailPayload`
   - `ArticleDetailProps` -> `TopicDetailProps`
   - `ReplyWindowPayload` -> `PostWindowPayload`
   - `PostPayload` 已是新名，优先沿用。
5. 完成后搜索确认：

```bash
rg "postReply|updateReply|deleteReply|articleId|replyId|replyNo|ArticlePayload|ArticleDetailProps|ReplyWindowPayload|ArticlePage|ArticleReplyComposer|ReplyPositionRail" resource/src
```

允许保留：

- i18n 文案中的“文章/回复”用户可见文本。
- 用户可见 URL/hash 相关的 `reply-` 需要单独评估是否迁移，避免破坏外链。
- 历史 payload 兼容字段。

## 阶段 2：后端运行时代码命名收敛

目标：后端新运行路径里的变量、请求结构、函数名尽量使用 `Topic/Post`，但不碰历史兼容和旧模型目录。

建议修改范围：

- `app/http/controllers/articleController.go`
- `app/http/controllers/forum/article.go`
- `app/http/controllers/forum/payload.go`
- `app/http/controllers/forum/moderation.go`
- `app/http/controllers/api/adminController.go`
- `app/service/eventhandlers/`
- `app/service/searchservice/`
- `app/service/fileusageservice/`

建议操作：

1. Controller request/response 内部类型改名：
   - `WriteArticleReq` -> `WriteTopicReq`
   - `ArticleStatusReq` -> `TopicStatusReq`
   - `UpdateReplyReq` -> `UpdatePostReq`
   - `DeletePostReq` 已是新名，沿用。
2. Handler 函数可逐步增加新名并替换路由引用：
   - `WriteArticles` -> `WriteTopic`
   - `UpdateArticleStatus` -> `UpdateTopicStatus`
   - `ArticleReply` -> `CreatePost`
   - `UpdateReply` -> `UpdatePost`
   - `DeleteReply` -> `DeletePost`
3. Admin handler 内部命名改为 Topic：
   - `ArticlesList` -> `TopicsList`
   - `ArticleSource` -> `TopicSource`
   - `EditArticle` -> `EditTopic`
   - `DeleteArticle` -> `DeleteTopic`
4. 日志 message key 与内部事件名也统一切到 topic/post，不继续保留 article/reply。
5. 完成后搜索确认：

```bash
rg "ArticleReply|WriteArticles|UpdateArticleStatus|ArticlesList|ArticleSource|EditArticle|DeleteArticle|ReplyId|ArticleId" app/http app/service
```

允许保留：

- 迁移代码里的旧字段。
- 日志、通知、举报历史兼容里的 `ArticleId` / `ReplyId`。
- 用户可见文案仍叫“文章/回复”的地方。

## 阶段 3：历史兼容边界整理

目标：确认迁移脚本已经覆盖旧字段，并清理运行时代码里的 `article/reply` 兼容分支，避免以后误以为还能继续写旧协议。

建议修改范围：

- `app/http/controllers/forum/moderation.go`
- `app/service/eventhandlers/`
- `app/service/notificationservice/`
- `app/service/moderationlogservice/`
- `app/models/forum/eventNotification/`

建议操作：

1. 迁移脚本集中处理旧字段：
   - 举报 `article/reply -> topic/post`。
   - 通知 payload `articleId/commentId -> topicId/postId`。
   - 审核日志 `subject/action/params` 从 `article/reply` 转成 `topic/post`。
2. 新写入只写 topic/post 字段。
3. 旧字段只允许出现在迁移脚本、迁移测试、deprecated model 或明确保留的历史存储字段里；模型层不再暴露 `reports.TargetArticle` / `reports.TargetReply`、`fileUsage.TargetArticle` / `fileUsage.TargetReply` 这类可供新代码复用的旧常量。
4. 为历史迁移补测试：
   - 旧 `reports.TargetArticle` / `reports.TargetReply` 能迁成 `topic/post`。
   - 旧通知 payload 能迁成 `topicId/postId`。
   - 旧审核日志 payload 能迁成 `topicId/postId/postNo`。
   - 旧 file usage target 能迁成 `topic/post`。
5. 完成后搜索确认：

```bash
rg "TargetArticle|TargetReply|articleId|commentId|replyId" app/service app/http/controllers/forum app/models/forum/reports
```

这里的目标是运行时代码尽量清零；残留应只在迁移脚本、迁移测试或用户可见文案里出现。

## 阶段 4：旧 model 运行时依赖清零

目标：确认运行时代码不再依赖旧 model 包，为删除旧目录做准备。

建议操作：

1. 搜索旧 model import：

```bash
rg "models/forum/articles|models/forum/reply|articleCategory|articleCategoryRs|articleUserAction|articlesUserStat" app --glob '*.go'
```

2. 分类处理：
   - 迁移代码允许暂时保留。
   - 测试旧迁移逻辑允许暂时保留。
   - 运行时代码不应再 import 旧 model。
3. 若还有运行时代码依赖旧 model，先替换为：
   - `topics`
   - `posts`
   - `category`
   - `topicCategoryIndex`
   - `topicUserAction`
   - `topicUserStat`
4. 增加 guard 测试，防止运行时重新引入旧 model。

建议验证：

```bash
go test ./app/migration ./app/http/... ./app/service/... ./app/models/... -count=1
go test ./...
```

## 阶段 5：旧 model 和旧表清理评估

目标：决定是否删除旧 model 目录，以及是否停止 AutoMigrate 旧表。

前置条件：

- 阶段 4 已确认运行时代码无旧 model import。
- 迁移脚本不再需要线上反复运行旧表读取，或者已经明确只作为一次性历史迁移保留。
- 有数据库备份和回滚策略。

建议操作：

1. 先从 AutoMigrate 清理旧 model。
2. 保留旧表，不立即 drop。
3. 发布一个版本观察。
4. 再单独评估是否提供手动清理命令，而不是启动时自动 drop。

不建议：

- 不建议在应用启动时自动删除旧表。
- 不建议把旧表删除和命名清理放在同一个版本。

## 推荐执行顺序

1. 前端内部命名收敛。
2. 后端运行时代码命名收敛。
3. 历史兼容边界整理和测试。
4. 旧 model 运行时依赖清零。
5. 旧 model 和旧表清理评估。

当前最适合下一步执行的是阶段 1。它风险最低，主要是 TypeScript 类型和调用名收敛；完成后再做后端函数/类型名会更顺。

# Topic Post Next Phase Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Move the remaining article/reply business surface toward the clean `topic/post` model while keeping current routes and user-facing behavior stable.

**Architecture:** Continue using migration-first, replacement-second. The new `topics`, `posts`, `category`, `topic_category_index`, `topic_user_action`, and `topic_user_stat` tables are the source of the next implementation slices; old `articles` and `reply` models stay only as compatibility inputs until each slice is replaced and verified.

**Tech Stack:** Go, Gin, GORM, SQLite/MySQL, Vue 3, TypeScript, Vite, TailwindCSS 4, localcache.

## Global Constraints

- Do not change public route paths in this phase; `/p/post/:id` and existing API endpoint paths stay compatible.
- Do not perform a broad mechanical rename before behavior has moved.
- Keep each commit independently testable.
- Keep cache invalidation explicit and scoped; avoid adding broad global clears unless existing behavior already does so.
- For `resource/` changes, run the required frontend build or dev workflow based on `config.toml`.
- Preserve old model imports only where the current compatibility layer still genuinely needs them.

---

## File Structure

- `app/models/forum/topics/`: topic read/write repository methods, counters, status, pin, list queries.
- `app/models/forum/posts/`: post read/write repository methods, reply windows, edit/delete/status helpers.
- `app/models/forum/topicCategoryIndex/`: topic-category relation sync helpers.
- `app/models/forum/topicUserAction/`: like/bookmark/watch repository methods.
- `app/models/forum/topicUserStat/`: per-user topic reply stats.
- `app/http/controllers/articleController.go`: current C-side write/action endpoints; replace internals while preserving request/response compatibility.
- `app/http/controllers/forum/article.go`: current C-side detail/reply-window read endpoint; move to topic/post reads.
- `app/http/controllers/forum/payload.go`: payload assembly; introduce topic/post builders, then retire article/reply builders.
- `app/http/controllers/forum/moderation.go`: already partially supports `topic/post`; finish once post status write path exists.
- `app/http/controllers/api/adminController.go`: admin article management; move to topic/post repositories and topic category index.
- `app/models/hotdataserve/article_list_cache.go`: rename behavior internally to topic list cache after reads move.
- `app/models/hotdataserve/category_cache.go`: switch from `articleCategory` to `category`.
- `app/service/moderationstatusservice/status.go`: invalidate by topic/category using new tables, not `articles`.
- `app/service/eventhandlers/`: events still named article/comment for compatibility; gradually add topic/post payload fields before renaming.
- `app/service/searchservice/`: rebuild index from topics and first posts.
- `resource/src/runtime/api.ts`, `resource/src/types/payload.ts`, `resource/src/site/pages/ArticlePage.vue`, admin pages: keep API compatibility first, then rename frontend types in a final cleanup.

---

### Task 1: Add Topic/Post Repository Parity

**Files:**
- Modify: `app/models/forum/topics/topics_rep.go`
- Modify: `app/models/forum/posts/posts_rep.go`
- Modify: `app/models/forum/topicCategoryIndex/topicCategoryIndex_rep.go`
- Modify: `app/models/forum/topicUserAction/topicUserAction_rep.go`
- Modify: `app/models/forum/topicUserStat/topicUserStat_rep.go`
- Test: `app/models/forum/topics/topics_test.go`
- Test: `app/models/forum/posts/posts_test.go`

**Interfaces:**
- Produces: topic repository methods matching current article capabilities: `Page`, `GetSimple`, `UpdateProcessStatus`, `UpdatePinWeight`, `IncrementLike`, `DecrementLike`, `ReservePostSequence`.
- Produces: post repository methods matching current reply capabilities: `GetFirstPageByTopicId`, `GetByTopicPostNoAsc`, `GetByTopicPostNoDesc`, `GetByTopicIdAfter`, `GetByTopicIdBefore`, `UpdateProcessStatus`, `GetMaxPostNoByTopicId`.

- [ ] **Step 1: Write failing repository tests**

Add tests that create topics/posts in an in-memory SQLite database and assert list filtering, category filtering, post windows, counter changes, and status updates work without importing `articles` or `reply`.

Run:

```bash
go test ./app/models/forum/topics ./app/models/forum/posts -count=1
```

Expected: FAIL because the repository methods do not exist or are incomplete.

- [ ] **Step 2: Implement minimal repository methods**

Add the methods needed by the tests. Use the existing query style from `articles_rep.go` and `reply_rep.go`, but target `topics` and `posts`.

- [ ] **Step 3: Verify repository tests**

Run:

```bash
go test ./app/models/forum/topics ./app/models/forum/posts -count=1
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add app/models/forum/topics app/models/forum/posts app/models/forum/topicCategoryIndex app/models/forum/topicUserAction app/models/forum/topicUserStat
git commit -m "feat: add topic post repository parity"
```

---

### Task 2: Move C-Side Read Path To Topics/Posts

**Files:**
- Modify: `app/http/controllers/forum/article.go`
- Modify: `app/http/controllers/forum/payload.go`
- Modify: `app/models/hotdataserve/article_list_cache.go`
- Modify: `app/models/hotdataserve/category_cache.go`
- Test: `app/http/controllers/forum/*_test.go`
- Test: `app/models/hotdataserve/*_test.go`

**Interfaces:**
- Consumes: Task 1 repository methods.
- Produces: existing page payload shape remains compatible: `ArticleDetailProps`, `ArticlePayload`, `ReplyPayload` can still be emitted while backed by `topics/posts`.

- [ ] **Step 1: Write failing read-path tests**

Add tests that seed only `topics`, `posts`, `category`, and `topic_category_index`, then call payload builders or page helpers. Assert the detail page and reply window can be assembled without old `articles/reply` rows.

Run:

```bash
go test ./app/http/controllers/forum ./app/models/hotdataserve -count=1
```

Expected: FAIL because current builders still load `articles/reply`.

- [ ] **Step 2: Add topic/post payload builders**

Implement topic-backed versions of the current article payload builders. Keep JSON field names stable for the frontend during this task.

- [ ] **Step 3: Switch list/detail read calls**

Update home/category/detail/reply-window read paths to use `topics/posts` and `topic_category_index`. Keep route and response names unchanged.

- [ ] **Step 4: Verify read path**

Run:

```bash
go test ./app/http/controllers/forum ./app/models/hotdataserve -count=1
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add app/http/controllers/forum app/models/hotdataserve
git commit -m "feat: read topics and posts on site pages"
```

---

### Task 3: Move C-Side Write Path To Topics/Posts

**Files:**
- Modify: `app/http/controllers/articleController.go`
- Modify: `app/service/replyservice/replyservice.go`
- Modify: `app/service/fileusageservice/file_usage_service.go`
- Modify: `app/service/eventhandlers/notification_handler.go`
- Modify: `app/service/eventhandlers/search_handler.go`
- Test: `app/http/controllers/*_test.go`
- Test: `app/service/replyservice/*_test.go`
- Test: `app/service/eventhandlers/*_test.go`

**Interfaces:**
- Consumes: Task 1 repository methods.
- Produces: create/update/delete topic and post behavior using `topics/posts`.
- Produces: existing request structs can stay compatible: `WriteArticleReq`, `ArticleReplyId`, `UpdateReplyReq`.

- [ ] **Step 1: Write failing write-path tests**

Add tests for topic creation, first-post creation, post reply creation, post edit, post delete, like/bookmark/watch, and event payload topic/post ids.

Run:

```bash
go test ./app/http/controllers ./app/service/replyservice ./app/service/eventhandlers -count=1
```

Expected: FAIL because current writes still use old tables.

- [ ] **Step 2: Replace topic creation/update internals**

In `WriteArticles`, create/update `topics` plus first `posts` row. Keep response as the topic id.

- [ ] **Step 3: Replace reply creation/update/delete internals**

Move `ArticleReply`, `UpdateReply`, and `DeleteReply` to `posts`. Map `ReplyId` compatibility to `ReplyToPostId` once frontend still sends old field names.

- [ ] **Step 4: Replace like/bookmark/watch internals**

Move `articleUserAction` usage to `topicUserAction`; update topic like counters.

- [ ] **Step 5: Verify write path**

Run:

```bash
go test ./app/http/controllers ./app/service/replyservice ./app/service/eventhandlers -count=1
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add app/http/controllers/articleController.go app/service/replyservice app/service/fileusageservice app/service/eventhandlers
git commit -m "feat: write topics and posts on site actions"
```

---

### Task 4: Move Admin Topic Management

**Files:**
- Modify: `app/http/controllers/api/adminController.go`
- Modify: `app/http/routes/route4api.go` only if handler names change internally; do not change endpoint paths.
- Modify: `resource/src/admin/pages/Articles*.vue`
- Modify: `resource/src/admin/api/*`
- Test: `app/http/controllers/api/*_test.go`

**Interfaces:**
- Consumes: Task 1 repository methods.
- Produces: admin list, source, status, delete, pin, and categories using topics/posts.
- Keeps existing admin endpoint paths until the frontend rename task.

- [ ] **Step 1: Write failing admin tests**

Add tests for admin status change, pin change, category sync, and delete using only new topic/category tables.

Run:

```bash
go test ./app/http/controllers/api -count=1
```

Expected: FAIL because admin controller still uses `articles` and `articleCategoryRs`.

- [ ] **Step 2: Switch admin controller internals**

Replace article model calls with topic/post repository calls. Replace category relation sync with `topicCategoryIndex`.

- [ ] **Step 3: Update admin frontend labels only where needed**

Keep endpoint paths stable. Change UI wording from article to topic/post only where the current page is explicitly管理内容/主题.

- [ ] **Step 4: Verify admin and frontend build**

Run:

```bash
go test ./app/http/controllers/api -count=1
go test ./...
pnpm -C resource build
```

Expected: Go tests PASS and frontend build exits 0.

- [ ] **Step 5: Commit**

```bash
git add app/http/controllers/api app/http/routes/route4api.go resource/src/admin
git commit -m "feat: manage topics through admin"
```

---

### Task 5: Rework Cache And Status Invalidations

**Files:**
- Modify: `app/models/hotdataserve/article_list_cache.go`
- Modify: `app/models/hotdataserve/category_cache.go`
- Modify: `app/models/hotdataserve/site_stats_cache.go`
- Modify: `app/service/moderationstatusservice/status.go`
- Modify: `app/service/unreadservice/unreadservice.go` only if topic/post notification scopes require it.
- Test: `app/models/hotdataserve/*_test.go`
- Test: `app/service/moderationstatusservice/*_test.go`

**Interfaces:**
- Consumes: Task 2 and Task 3 topic/post read/write paths.
- Produces: cache invalidation by topic id and category id.
- Produces: moderation red dot cache using topic category scope.

- [ ] **Step 1: Write failing cache tests**

Add tests for category cache loading from `category`, topic list cache invalidation, and moderation status invalidation from topic category ids.

Run:

```bash
go test ./app/models/hotdataserve ./app/service/moderationstatusservice -count=1
```

Expected: FAIL because cache loaders or invalidation still use old models.

- [ ] **Step 2: Switch category cache to `category`**

Replace `articleCategory` imports and mapping with clean `category` model output.

- [ ] **Step 3: Rename cache functions internally**

Introduce `ClearTopicListCache`, `GetLatestTopicsSimpleVoPaginated`, and `GetTopicsByCategorySimpleVo`. Keep old function wrappers temporarily if callers still exist.

- [ ] **Step 4: Switch moderation status invalidation**

Change `InvalidateArticle` into `InvalidateTopic`, keep a wrapper during compatibility, and load categories from `topics` or `topic_category_index`.

- [ ] **Step 5: Verify cache behavior**

Run:

```bash
go test ./app/models/hotdataserve ./app/service/moderationstatusservice -count=1
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add app/models/hotdataserve app/service/moderationstatusservice app/service/unreadservice
git commit -m "feat: cache topic and category data"
```

---

### Task 6: Update Search, Notifications, And HTTP Notify Naming

**Files:**
- Modify: `app/service/searchservice/indexservice.go`
- Modify: `app/service/searchservice/struct.go`
- Modify: `app/service/eventhandlers/http_notify_handler.go`
- Modify: `app/service/eventhandlers/notification_handler.go`
- Modify: `app/service/eventnotice/eventNotice.go`
- Modify: `app/models/forum/eventNotification/eventNotification.go`
- Test: `app/service/searchservice/*_test.go`
- Test: `app/service/eventhandlers/*_test.go`
- Test: `app/service/eventnotice/*_test.go`

**Interfaces:**
- Consumes: Task 3 events carrying topic/post ids.
- Produces: search index documents built from topics and first posts.
- Produces: notification payloads that prefer `topicId/postId`; legacy JSON fields stay only for backward compatibility if tests require them.

- [ ] **Step 1: Write failing service tests**

Add tests for topic search document conversion, comment notification, reply notification, watcher notification, and HTTP notify payload with topic/post ids.

Run:

```bash
go test ./app/service/searchservice ./app/service/eventhandlers ./app/service/eventnotice -count=1
```

Expected: FAIL because several services still load old article/reply rows.

- [ ] **Step 2: Switch search document source**

Build search documents from topic plus first post content. Keep index name stable unless a separate deployment migration is planned.

- [ ] **Step 3: Switch notification event internals**

Use topic/post ids in event handlers and notification service calls.

- [ ] **Step 4: Switch HTTP notify payload internals**

Keep external JSON compatible where useful, but derive payload from topic/post rows.

- [ ] **Step 5: Verify services**

Run:

```bash
go test ./app/service/searchservice ./app/service/eventhandlers ./app/service/eventnotice -count=1
go test ./...
```

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add app/service/searchservice app/service/eventhandlers app/service/eventnotice app/models/forum/eventNotification
git commit -m "feat: notify and search with topic post data"
```

---

### Task 7: Frontend Compatibility Rename Pass

**Files:**
- Modify: `resource/src/types/payload.ts`
- Modify: `resource/src/runtime/api.ts`
- Modify: `resource/src/site/pages/ArticlePage.vue`
- Modify: `resource/src/site/pages/NotificationsPage.vue`
- Modify: `resource/src/site/components/*Article*.vue`
- Modify: `resource/src/admin/**/*Article*.vue`
- Modify: `resource/src/locales/*.ts`

**Interfaces:**
- Consumes: backend-compatible payloads from Tasks 2 through 6.
- Produces: frontend names move toward topic/post while keeping rendered UI stable.

- [ ] **Step 1: Write frontend type check target**

Run the existing frontend build first to capture baseline:

```bash
pnpm -C resource build
```

Expected: PASS before edits.

- [ ] **Step 2: Rename TypeScript types gradually**

Introduce `TopicPayload` and `PostPayload` aliases first. Move call sites from `ArticlePayload` and `ReplyPayload` where no backend JSON field change is needed.

- [ ] **Step 3: Rename UI component internals**

Rename local variables and component names only where it improves clarity. Keep translation keys stable unless a key is obviously internal-only.

- [ ] **Step 4: Verify frontend**

Run:

```bash
pnpm -C resource build
go test ./...
```

Expected: frontend build exits 0 and Go tests PASS.

- [ ] **Step 5: Commit**

```bash
git add resource/src
git commit -m "refactor: rename frontend topic post types"
```

---

### Task 8: Remove Old Model Dependencies From Active Runtime

**Files:**
- Modify: active Go files under `app/http`, `app/service`, `app/models/hotdataserve`, and `app/models/forum`.
- Keep: old model directories only if a final compatibility migration still imports them nowhere.
- Test: `app/migration/migration_test.go`
- Test: new dependency guard tests as needed.

**Interfaces:**
- Consumes: all previous tasks.
- Produces: active runtime no longer imports old `articles`, `reply`, `articleCategory`, `articleCategoryRs`, `articleUserAction`, or `articlesUserStat`.

- [ ] **Step 1: Add dependency guard tests**

Extend existing migration/runtime guard tests to fail if active runtime packages import old article/reply models after this phase.

Run:

```bash
go test ./app/migration ./app/http/... ./app/service/... -count=1
```

Expected: FAIL while old imports remain.

- [ ] **Step 2: Remove remaining active imports**

Replace or delete old imports from active runtime code. Leave legacy data migration code free of old model imports, matching the current migration direction.

- [ ] **Step 3: Verify complete backend**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add app
git commit -m "refactor: remove active article reply model usage"
```

---

## Current Risk Notes

- Reports are now migrated to `topic/post`, but report creation still accepts old request target names for page compatibility. That should be revisited during Task 3 or Task 7.
- Notification payloads have moved toward `topicId/postId`, but event names and several service structs still use article/comment wording.
- Admin topic management is the riskiest non-public surface because it mixes status, deletion, pinning, category sync, search rebuild, opt logs, and moderation logs.
- Cache migration should not be delayed until the final cleanup; stale list/category/moderation cache can hide correctness problems.
- Search index naming can remain `articles` for deployment stability; rename the external index only in a separate operational plan.

## Self-Review

- Spec coverage: backend, admin, cache, services, frontend compatibility, and old dependency cleanup are each covered by a separate task.
- Placeholder scan: this plan avoids open-ended markers and gives exact file areas and verification commands for every task.
- Type consistency: the plan consistently treats `topic` as the old article id during route compatibility, and `post` as the new row id generated by migration.

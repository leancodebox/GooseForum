# Topic/Post Stream 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将主题详情和帖子窗口统一到 `TopicMeta + PostStream` 结构，首楼作为 `postNo = 1` 的 `PostPayload` 返回并由前端统一渲染。

**Architecture:** 后端抽出共享 post stream 构造逻辑，`TopicDetail` 使用它生成首屏 stream，`PostWindow` 使用它生成后续窗口。前端 `TopicPage` 从 `postStream.posts` 初始化并渲染所有楼层，首楼通过 `post.postNo === 1` 保留 topic 级操作。

**Tech Stack:** Go + Gin + GORM，Vue 3 + TypeScript + Vite，现有 payload/runtime API 结构。

## Global Constraints

- 不修改公开路由。
- 不修改数据库存储结构。
- 不引入新的第三方依赖。
- 首楼正文不再通过 `topic.html` 暴露。
- `PostPayload` 必须支持 `postNo = 1`。
- 时间轴、跳楼、删除楼层空洞继续使用真实 `postNo`。
- 每个任务完成后运行对应测试；最终运行 `go test ./...`、`cd resource && npx vitest run`、`cd resource && pnpm build`、`git diff --check`。

---

## File Structure

- Modify: `app/http/controllers/forum/payload.go`
  - 更新 `TopicDetailProps`、`TopicDetailPayload`。
  - 让 `buildPostPayloads` 支持首楼。
  - 新增或调整 post stream payload 构造辅助函数。

- Modify: `app/http/controllers/forum/topic.go`
  - 让 `TopicDetail` 和 `PostWindow` 复用统一 post stream 逻辑。
  - 移除 `filterSecondaryPosts`。
  - 允许 `PostWindow` 锚定 `postNo = 1`。

- Modify: `app/http/controllers/forum/topic_meta_test.go`
  - 更新 topic detail、post window、删除楼层空洞相关测试。

- Modify: `resource/src/types/payload.ts`
  - `TopicDetailProps.posts` 改为 `TopicDetailProps.postStream`。
  - 移除 `TopicDetailPayload.html`。

- Modify: `resource/src/site/pages/TopicPage.vue`
  - 从 `page.props.postStream.posts` 初始化。
  - 删除单独渲染 `page.props.topic.html` 的正文块。
  - 在统一 post 列表里对 `post.postNo === 1` 做首楼特殊展示和操作。

- Modify: `resource/src/runtime/api.ts`
  - 如需调整 `PostWindowPayload` 类型使用，保持 `getPostWindow` 返回结构不变。

---

### Task 1: 后端 payload 支持首楼 PostPayload

**Files:**
- Modify: `app/http/controllers/forum/payload.go`
- Test: `app/http/controllers/forum/topic_meta_test.go`

**Interfaces:**
- Consumes: existing `PostPayload`, `TopicDetailProps`, `TopicDetailPayload`, `PostWindowPayload`
- Produces:
  - `TopicDetailProps.PostStream PostWindowPayload`
  - `buildPostPayloads(postEntities []*posts.Entity, userMap map[uint64]*users.EntityComplete, currentUserID uint64, canModerate bool) []PostPayload` supports `PostNo == 1`
  - `TopicDetailPayload` no longer contains `HTML string`

- [ ] **Step 1: Update failing tests**

Update `TestBuildTopicDetailPropsReadsTopicPostTables` so it expects the first post inside `props.PostStream.Posts`.

Expected assertions:

```go
if props.Topic.ID != topicID || props.Topic.MaxPostNo != 2 {
	t.Fatalf("topic payload mismatch: %#v", props.Topic)
}
if len(props.PostStream.Posts) != 2 {
	t.Fatalf("post stream length = %d, want 2", len(props.PostStream.Posts))
}
if props.PostStream.Posts[0].ID != firstPostID || props.PostStream.Posts[0].PostNo != 1 || props.PostStream.Posts[0].RenderedContent != "<p>first</p>" {
	t.Fatalf("first post payload mismatch: %#v", props.PostStream.Posts[0])
}
if props.PostStream.Posts[1].ID != replyPostID || props.PostStream.Posts[1].PostNo != 2 {
	t.Fatalf("reply post payload mismatch: %#v", props.PostStream.Posts[1])
}
if props.PostStream.BeforePostNo != 1 || props.PostStream.AfterPostNo != 2 || props.PostStream.MaxPostNo != 2 {
	t.Fatalf("post stream cursor mismatch: %#v", props.PostStream)
}
```

Update `TestTopicMetaJSONLDIncludesForumRequiredFields` and `TestTopicMetaJSONLDIncludesImageForImageOnlyTopic` to stop setting `HTML`. Keep `Description` and `FirstImageURL` assertions.

- [ ] **Step 2: Run backend test to verify failure**

Run:

```bash
go test ./app/http/controllers/forum -run 'TestBuildTopicDetailPropsReadsTopicPostTables|TestTopicMetaJSONLDIncludesForumRequiredFields|TestTopicMetaJSONLDIncludesImageForImageOnlyTopic' -count=1
```

Expected: compile failure or assertion failure because `PostStream` does not exist and `TopicDetailPayload.HTML` still exists.

- [ ] **Step 3: Update payload structs**

In `app/http/controllers/forum/payload.go`, change:

```go
type TopicDetailProps struct {
	Topic       TopicDetailPayload `json:"topic"`
	PostStream  PostWindowPayload  `json:"postStream"`
	HotTopics   []TopicPayload     `json:"hotTopics"`
	Permissions TopicPermissions   `json:"permissions"`
}
```

Remove this field from `TopicDetailPayload`:

```go
HTML string `json:"html"`
```

- [ ] **Step 4: Allow buildPostPayloads to include first post**

In `buildPostPayloads`, replace:

```go
if item == nil || item.PostNo <= 1 {
	continue
}
```

with:

```go
if item == nil {
	continue
}
```

Keep existing hidden-content behavior unchanged.

- [ ] **Step 5: Add a local post stream payload helper**

In `app/http/controllers/forum/payload.go`, add:

```go
func buildPostWindowPayloadFromEntities(postEntities []*posts.Entity, userMap map[uint64]*users.EntityComplete, currentUserID uint64, canModerate bool, hasBefore bool, hasAfter bool, total int64, maxPostNo uint64, anchorPostID uint64) PostWindowPayload {
	payloadPosts := buildPostPayloads(postEntities, userMap, currentUserID, canModerate)
	var beforeCursor uint64
	var afterCursor uint64
	var beforePostNo uint64
	var afterPostNo uint64
	if len(postEntities) > 0 {
		beforeCursor = postEntities[0].Id
		afterCursor = postEntities[len(postEntities)-1].Id
		beforePostNo = postEntities[0].PostNo
		afterPostNo = postEntities[len(postEntities)-1].PostNo
	}
	return PostWindowPayload{
		Posts:        payloadPosts,
		AnchorPostID: anchorPostID,
		BeforeCursor: beforeCursor,
		AfterCursor:  afterCursor,
		BeforePostNo: beforePostNo,
		AfterPostNo:  afterPostNo,
		HasBefore:    hasBefore,
		HasAfter:     hasAfter,
		Total:        total,
		MaxPostNo:    maxPostNo,
	}
}
```

- [ ] **Step 6: Update buildTopicDetailPayload**

Remove first-post HTML handling from `buildTopicDetailPayload`.

Keep `CreatedAt` and `UpdatedAt` derived from `firstPost`:

```go
createdAt := topic.CreatedAt
updatedAt := topic.UpdatedAt
if firstPost != nil {
	createdAt = firstPost.CreatedAt
	updatedAt = firstPost.UpdatedAt
}
```

Do not set an `HTML` field in the returned struct.

- [ ] **Step 7: Update buildTopicDetailProps**

Make `postEntities` include the first post followed by the first page replies:

```go
replyEntities := posts.GetFirstPageByTopicId(topic.Id)
postEntities := make([]*posts.Entity, 0, 1+len(replyEntities))
if firstPost != nil && firstPost.Id != 0 {
	postEntities = append(postEntities, firstPost)
}
postEntities = append(postEntities, replyEntities...)
```

Build users from `postEntities`.

Return:

```go
PostStream: buildPostWindowPayloadFromEntities(
	postEntities,
	userMap,
	currentUserID,
	moderatorservice.CanModerateAnyCategory(currentUserID, topic.CategoryIds),
	false,
	topic.ReplyCount > uint64(len(replyEntities)),
	int64(topic.PostSeq),
	topic.PostSeq,
	0,
),
```

Use `Total: int64(topic.PostSeq)` because the stream includes the first post.

- [ ] **Step 8: Run backend test**

Run:

```bash
go test ./app/http/controllers/forum -run 'TestBuildTopicDetailPropsReadsTopicPostTables|TestTopicMetaJSONLDIncludesForumRequiredFields|TestTopicMetaJSONLDIncludesImageForImageOnlyTopic' -count=1
```

Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add app/http/controllers/forum/payload.go app/http/controllers/forum/topic_meta_test.go
git commit -m "refactor: expose topic detail post stream"
```

---

### Task 2: 后端 PostWindow 统一支持完整 PostStream

**Files:**
- Modify: `app/http/controllers/forum/topic.go`
- Modify: `app/http/controllers/forum/topic_meta_test.go`

**Interfaces:**
- Consumes:
  - `buildPostWindowPayloadFromEntities(...) PostWindowPayload`
  - `buildPostPayloads(...)` supports `postNo = 1`
- Produces:
  - `PostWindow` accepts `AnchorPostNo = 1`
  - `PostWindow` no longer filters out first post
  - `PostWindowPayload.Total` is stream total, not reply-only total

- [ ] **Step 1: Update failing tests**

Rename `TestPostWindowSkipsFirstPostInCursors` to `TestPostWindowTailIncludesFirstPostWhenWindowCoversWholeStream`.

Change expectations:

```go
if len(payload.Posts) != 2 || payload.Posts[0].ID != firstPostID || payload.Posts[0].PostNo != 1 || payload.Posts[1].ID != replyPostID || payload.Posts[1].PostNo != 2 {
	t.Fatalf("posts = %#v, want first post and reply post", payload.Posts)
}
if payload.BeforeCursor != firstPostID || payload.AfterCursor != replyPostID || payload.BeforePostNo != 1 || payload.AfterPostNo != 2 {
	t.Fatalf("cursor payload = %#v", payload)
}
if payload.Total != 2 || payload.MaxPostNo != 2 {
	t.Fatalf("stream totals = %#v", payload)
}
```

Add a new test `TestPostWindowAnchorPostNoCanLoadFirstPost`:

```go
func TestPostWindowAnchorPostNoCanLoadFirstPost(t *testing.T) {
	preferences.Set("db.default.connection", "sqlite")
	preferences.Set("db.default.path", ":memory:")
	conn := dbconnect.Connect()
	if err := conn.AutoMigrate(&topics.Entity{}, &posts.Entity{}, &users.EntityComplete{}); err != nil {
		t.Fatalf("migrate post window tables: %v", err)
	}

	now := time.Date(2026, 7, 7, 14, 0, 0, 0, time.UTC)
	topicID := uint64(993010)
	firstPostID := uint64(993100)
	replyPostID := uint64(993101)
	userID := uint64(993001)
	t.Cleanup(func() {
		conn.Unscoped().Where("topic_id = ?", topicID).Delete(&posts.Entity{})
		conn.Unscoped().Delete(&topics.Entity{}, topicID)
		conn.Unscoped().Delete(&users.EntityComplete{}, userID)
	})
	conn.Create(&users.EntityComplete{Id: userID, Username: "author"})
	conn.Create(&topics.Entity{Id: topicID, Title: "topic", UserId: userID, Status: 1, ReplyCount: 1, PostSeq: 2, FirstPostId: firstPostID, CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: firstPostID, TopicId: topicID, PostNo: 1, UserId: userID, Content: "first", CreatedAt: now, UpdatedAt: now})
	conn.Create(&posts.Entity{Id: replyPostID, TopicId: topicID, PostNo: 2, UserId: userID, Content: "reply", CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute)})

	res := PostWindow(component.BetterRequest[PostWindowReq]{
		Params: PostWindowReq{TopicID: topicID, AnchorPostNo: 1, Limit: 20},
	})
	payload, ok := res.Data.Result.(PostWindowPayload)
	if !ok {
		t.Fatalf("result type = %T, want PostWindowPayload", res.Data.Result)
	}
	if len(payload.Posts) != 2 || payload.Posts[0].PostNo != 1 || payload.Posts[1].PostNo != 2 {
		t.Fatalf("payload posts = %#v, want first post and reply", payload.Posts)
	}
	if payload.HasBefore {
		t.Fatalf("first post anchor should not have before window: %#v", payload)
	}
}
```

- [ ] **Step 2: Run backend test to verify failure**

Run:

```bash
go test ./app/http/controllers/forum -run 'TestPostWindowTailIncludesFirstPostWhenWindowCoversWholeStream|TestPostWindowAnchorPostNoCanLoadFirstPost|TestPostWindowAnchorPostNoFallsForwardAcrossDeletedReplies' -count=1
```

Expected: failure because `PostWindow` still rejects or filters first post.

- [ ] **Step 3: Remove first-post rejection and filter**

In `app/http/controllers/forum/topic.go`, change these checks:

```go
anchor.PostNo <= 1
```

to:

```go
anchor.PostNo < 1
```

Delete this line:

```go
postEntities = filterSecondaryPosts(postEntities)
```

Delete the `filterSecondaryPosts` function.

- [ ] **Step 4: Make default PostWindow start from first post**

Change the default branch from:

```go
postEntities = posts.GetByTopicPostNoAfter(topicID, 1, limit+1)
```

to:

```go
postEntities = posts.GetByTopicPostNoAfter(topicID, 0, limit+1)
```

This lets the API represent a full stream window by default.

- [ ] **Step 5: Use shared payload helper and stream totals**

Replace the manual payload assembly in `PostWindow` with:

```go
return component.SuccessResponse(buildPostWindowPayloadFromEntities(
	postEntities,
	userMap,
	req.UserId,
	canModeratePosts,
	hasBefore,
	hasAfter,
	int64(maxPostNo),
	maxPostNo,
	req.Params.AnchorPostID,
))
```

Keep `maxPostNo` fallback logic.

- [ ] **Step 6: Run backend tests**

Run:

```bash
go test ./app/http/controllers/forum -run 'TestPostWindowTailIncludesFirstPostWhenWindowCoversWholeStream|TestPostWindowAnchorPostNoCanLoadFirstPost|TestPostWindowAnchorPostNoFallsForwardAcrossDeletedReplies' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add app/http/controllers/forum/topic.go app/http/controllers/forum/topic_meta_test.go
git commit -m "refactor: unify post window stream"
```

---

### Task 3: 前端类型切换到 postStream

**Files:**
- Modify: `resource/src/types/payload.ts`
- Modify: `resource/src/site/pages/TopicPage.vue`

**Interfaces:**
- Consumes:
  - `TopicDetailProps.postStream: PostWindowPayload`
  - `TopicDetailPayload` without `html`
- Produces:
  - TypeScript compiles without `page.props.posts` and `page.props.topic.html`

- [ ] **Step 1: Update TypeScript payload types**

In `resource/src/types/payload.ts`, change `TopicDetailProps`:

```ts
export interface TopicDetailProps {
  topic: TopicDetailPayload
  postStream: PostWindowPayload
  hotTopics: TopicPayload[]
  permissions: {
    isOwnTopic: boolean
    canPost: boolean
    canModerateTopic: boolean
  }
}
```

Remove from `TopicDetailPayload`:

```ts
html: string
```

- [ ] **Step 2: Add local aliases in TopicPage**

In `resource/src/site/pages/TopicPage.vue`, add after `useFlashMessages()`:

```ts
const initialPostStream = page.props.postStream
const initialPosts = initialPostStream.posts
```

Replace initial reads:

```ts
const posts = ref<PostPayload[]>([...initialPosts])
const postHasBefore = ref(initialPostStream.hasBefore)
const postHasAfter = ref(initialPostStream.hasAfter)
const postBeforeCursor = ref(initialPostStream.beforeCursor || firstPostId(initialPosts))
const postAfterCursor = ref(initialPostStream.afterCursor || lastPostId(initialPosts))
const postBeforePostNo = ref(initialPostStream.beforePostNo || firstPostNo(initialPosts))
const postAfterPostNo = ref(initialPostStream.afterPostNo || lastPostNo(initialPosts))
const postMaxNo = ref(initialPostStream.maxPostNo || initialMaxPostNo())
const postTailLoaded = ref(!initialPostStream.hasAfter)
const activePostNo = ref(firstPostNo(initialPosts) || 1)
```

- [ ] **Step 3: Replace resetPageState references**

In `resetPageState`, replace `page.props.posts` with `initialPosts` and `page.props.postStream` fields:

```ts
posts.value = [...initialPosts]
postHasBefore.value = initialPostStream.hasBefore
postHasAfter.value = initialPostStream.hasAfter
postBeforeCursor.value = initialPostStream.beforeCursor || firstPostId(initialPosts)
postAfterCursor.value = initialPostStream.afterCursor || lastPostId(initialPosts)
postBeforePostNo.value = initialPostStream.beforePostNo || firstPostNo(initialPosts)
postAfterPostNo.value = initialPostStream.afterPostNo || lastPostNo(initialPosts)
postMaxNo.value = initialPostStream.maxPostNo || initialMaxPostNo()
postTailLoaded.value = !initialPostStream.hasAfter
activePostNo.value = firstPostNo(initialPosts) || 1
```

- [ ] **Step 4: Update helper functions**

Change `initialMaxPostNo`:

```ts
function initialMaxPostNo() {
  return Math.max(page.props.topic.maxPostNo || 0, page.props.postStream.maxPostNo || 0, lastPostNo(initialPosts))
}
```

Change `hasMoreInitialReplies` to stream semantics:

```ts
function hasMoreInitialReplies() {
  return page.props.postStream.hasAfter
}
```

- [ ] **Step 5: Run TypeScript/build to verify expected template errors remain**

Run:

```bash
cd resource && pnpm build
```

Expected: failure if template still references `page.props.topic.html`, or PASS if previous steps replaced all type references. If failure mentions `topic.html`, continue Task 4.

---

### Task 4: 前端统一渲染首楼和回复

**Files:**
- Modify: `resource/src/site/pages/TopicPage.vue`

**Interfaces:**
- Consumes:
  - `posts: Ref<PostPayload[]>` includes first post
  - `post.postNo === 1` identifies topic body
- Produces:
  - No `page.props.topic.html` usage
  - One post list renders both first post and replies

- [ ] **Step 1: Add first-post helpers**

In `TopicPage.vue` script section, add:

```ts
function isFirstPost(post: PostPayload) {
  return post.postNo === 1
}

function canEditPost(post: PostPayload) {
  return post.isOwnPost && !post.isHidden
}

function canDeleteRenderedPost(post: PostPayload) {
  return post.isOwnPost && !post.isHidden && !isFirstPost(post)
}
```

- [ ] **Step 2: Remove standalone topic body prose**

Delete the standalone block that renders:

```vue
<div class="gf-prose gf-prose-post" v-html="page.props.topic.html" />
```

Keep the topic header, title, categories, stats, and topic action buttons around it.

- [ ] **Step 3: Adjust post list template conditions**

In the `v-for="post in posts"` article controls:

Replace edit condition:

```vue
v-if="post.isOwnPost && !post.isHidden"
```

with:

```vue
v-if="canEditPost(post)"
```

Replace delete condition:

```vue
v-if="post.isOwnPost && !post.isHidden"
```

with:

```vue
v-if="canDeleteRenderedPost(post)"
```

For moderation buttons, keep post moderation for replies:

```vue
v-if="!isFirstPost(post) && post.canModerate && post.processStatus === 0"
```

and:

```vue
v-else-if="!isFirstPost(post) && post.canModerate && post.processStatus === 1"
```

Topic-level moderation remains in the topic header.

- [ ] **Step 4: Add first-post visual metadata**

Inside the post author/name row, add a small first-post label:

```vue
<span v-if="isFirstPost(post)" class="rounded bg-base-200 px-1.5 py-0.5 text-xs font-semibold text-base-content/55">
  {{ t('topic.originalPost') }}
</span>
```

Add i18n key in all locale files that already contain `topic.reply` keys:

```json
"originalPost": "正文"
```

Use equivalent translations:

- English: `"Original post"`
- Japanese: `"本文"`

- [ ] **Step 5: Ensure load-before is not shown above first post**

When initial stream includes first post, `postHasBefore` should be false. In `applyPostWindowPayload`, keep payload values authoritative:

```ts
postHasBefore.value = payload.hasBefore
postHasAfter.value = payload.hasAfter
```

Do not recompute before-state from reply count.

- [ ] **Step 6: Run frontend checks**

Run:

```bash
cd resource && npx vitest run
cd resource && pnpm build
```

Expected: both exit 0.

- [ ] **Step 7: Commit**

```bash
git add resource/src/types/payload.ts resource/src/site/pages/TopicPage.vue resource/src/**/*.json
git commit -m "refactor: render topic detail from post stream"
```

---

### Task 5: 删除旧接口残留并全量验证

**Files:**
- Modify: `app/http/controllers/forum/payload.go`
- Modify: `app/http/controllers/forum/topic.go`
- Modify: `resource/src/site/pages/TopicPage.vue`
- Modify: `resource/src/types/payload.ts`
- Test: existing backend/frontend tests

**Interfaces:**
- Consumes:
  - completed backend `postStream`
  - completed frontend unified render
- Produces:
  - no `topic.html` detail-body usage
  - no `filterSecondaryPosts`
  - no hidden first-post exclusion in forum detail/window path

- [ ] **Step 1: Search old references**

Run:

```bash
rg -n "topic\\.html|TopicDetailPayload.*HTML|json:\"html\"|page\\.props\\.posts|filterSecondaryPosts|PostNo <= 1|postNo <= 1" app resource/src -S
```

Expected allowed remaining matches:

- `PostNo <= 1` in write/edit/delete APIs that intentionally protect first post from reply-only operations.
- `postNo <= 1` in frontend jump logic if it intentionally routes to top of topic for post 1.

No allowed matches:

- `page.props.topic.html`
- `page.props.posts`
- `filterSecondaryPosts`
- `TopicDetailPayload.HTML`

- [ ] **Step 2: Fix any disallowed reference**

For each disallowed reference:

- Replace detail body usage with `postStream.posts`.
- Replace `page.props.posts` with `page.props.postStream.posts`.
- Delete unused helper functions.
- Keep reply-only guards in `app/http/controllers/topicController.go` unchanged.

- [ ] **Step 3: Run backend tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 4: Run frontend tests**

Run:

```bash
cd resource && npx vitest run
```

Expected: PASS.

- [ ] **Step 5: Run frontend build**

Run:

```bash
cd resource && pnpm build
```

Expected: exit 0. Existing Rolldown annotation warning from dependencies is acceptable if build exits 0.

- [ ] **Step 6: Run diff check**

Run:

```bash
git diff --check
```

Expected: no output, exit 0.

- [ ] **Step 7: Commit**

```bash
git add app/http/controllers/forum/payload.go app/http/controllers/forum/topic.go app/http/controllers/forum/topic_meta_test.go resource/src/types/payload.ts resource/src/site/pages/TopicPage.vue resource/src
git commit -m "refactor: complete topic post stream detail"
```

---

## Self-Review

- Spec coverage: The plan covers server post stream unification, payload changes, first post as `PostPayload`, frontend unified rendering, compatibility cleanup, and testing.
- Placeholder scan: No task uses open-ended implementation placeholders; each task names files, interfaces, commands, and expected outcomes.
- Type consistency: The plan consistently uses `TopicDetailProps.PostStream` in Go and `TopicDetailProps.postStream` in TypeScript/JSON, and keeps `PostWindowPayload` as the shared stream payload.

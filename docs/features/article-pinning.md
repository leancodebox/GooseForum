# Article Pinning Plan

## Goal

Add an admin-operated article pinning feature so selected published articles can stay above normal article lists without corrupting ordinary activity ordering.

The feature should support:

- Pinning and unpinning from admin post management.
- Optional category-scoped pinning.
- Optional expiration.
- Stable ordering among multiple pinned articles.
- Clear visual labels on customer-facing topic lists.

## Current Code Baseline

Article data is stored in `app/models/forum/articles`.

Important current fields:

- `article_status`: draft/published state. `1` means published.
- `process_status`: moderation state. `0` means normal, `1` means blocked.
- `category_id`: JSON category id list.
- `reply_count`, `view_count`, `like_count`.
- `created_at`, `updated_at`.

Current list reads:

- Home uses `hotdataserve.GetLatestArticlesSimpleVoPaginated(page, sort)`.
- Category pages use `hotdataserve.GetArticlesByCategorySimpleVo(id, sort, page)`.
- Admin post management uses `/api/admin/articles-list`, backed by `articles.Page`.

Current admin operations:

- `/api/admin/articles-list`: list posts.
- `/api/admin/article-source`: view source.
- `/api/admin/article-edit`: change `process_status`.
- `/api/admin/article-categories-edit`: change categories.

There is no article pinning field or operation today.

## Recommended Data Model

Add explicit pinning fields to `articles`:

```sql
ALTER TABLE articles
  ADD COLUMN pin_scope varchar(32) NOT NULL DEFAULT '' COMMENT '置顶范围：空=不置顶，global=全站，category=分类内',
  ADD COLUMN pin_order int NOT NULL DEFAULT 0 COMMENT '置顶排序，越大越靠前',
  ADD COLUMN pinned_at datetime NULL COMMENT '置顶时间',
  ADD COLUMN pinned_until datetime NULL COMMENT '置顶过期时间';
```

Recommended indexes:

```sql
CREATE INDEX idx_articles_pin_global
  ON articles (pin_scope, pin_order, pinned_at, article_status, process_status);

CREATE INDEX idx_articles_pin_until
  ON articles (pinned_until);
```

Notes:

- `pin_scope = ''`: not pinned.
- `pin_scope = 'global'`: pinned on the home/global list.
- `pin_scope = 'category'`: pinned only within categories the article already belongs to.
- `pin_order`: manual priority. Higher value appears first.
- `pinned_until IS NULL`: no expiration.
- Expired pins should be ignored by queries. A cleanup job can clear them later, but query correctness must not depend on cleanup.

## Why Not Reuse `updated_at`

Do not implement pinning by bumping `updated_at`.

Reasons:

- `updated_at` currently represents activity freshness and reply activity.
- Bumping it for pinning changes latest-reply ordering semantics.
- It can pollute RSS, search index freshness, hot topic sidebars, and user-facing “last update” display.
- Unpinning becomes lossy because the original activity time is gone unless extra state is added.

Explicit fields are slightly more work but much safer.

## Query Rules

Pinned articles must still obey visibility filters:

- `article_status = 1`
- `process_status = 0`
- `deleted_at IS NULL`
- `pinned_until IS NULL OR pinned_until > NOW()`

Home/global list recommended ordering:

```sql
ORDER BY
  CASE WHEN pin_scope = 'global'
        AND (pinned_until IS NULL OR pinned_until > NOW())
       THEN 1 ELSE 0 END DESC,
  pin_order DESC,
  pinned_at DESC,
  updated_at DESC
```

Category list recommended ordering:

```sql
ORDER BY
  CASE WHEN pin_scope IN ('global', 'category')
        AND (pinned_until IS NULL OR pinned_until > NOW())
       THEN 1 ELSE 0 END DESC,
  pin_order DESC,
  pinned_at DESC,
  updated_at DESC
```

For category pages, `pin_scope = 'category'` only takes effect if the article is in that category through `article_category_rs`.

## API Design

Add admin endpoint:

```text
POST /api/admin/article-pin
```

Request:

```json
{
  "id": 123,
  "pinScope": "global",
  "pinOrder": 100,
  "pinnedUntil": "2026-06-30 23:59:59"
}
```

Unpin request:

```json
{
  "id": 123,
  "pinScope": "",
  "pinOrder": 0,
  "pinnedUntil": ""
}
```

Validation:

- `id` is required.
- Article must exist.
- `pinScope` must be one of `""`, `"global"`, `"category"`.
- `pinOrder` should be bounded, for example `0..9999`.
- `pinnedUntil`, when present, must parse as local time and be in the future.
- Operation requires the same permission group currently used by article management, `permission.ArticlesManager`.

Response:

```json
true
```

Implementation location:

- Controller: `app/http/controllers/api/adminController.go`
- Route: `app/http/routes/route4api.go`
- Admin API wrapper: `resource/src/admin/runtime/api.ts`
- Admin type: `resource/src/admin/types.ts`

## Admin Operation Flow

In `resource/src/admin/pages/management/PostsManagementPage.vue`:

1. Add pin state display to each row:
   - Not pinned.
   - Global pinned.
   - Category pinned.
   - Expired pinned state, if the DB still contains expired pin metadata.
2. Add an action button, for example “置顶”.
3. Open a dialog with:
   - Scope segmented control: `不置顶 / 全站置顶 / 分类置顶`.
   - Priority numeric input.
   - Optional expiration datetime.
4. Save through `/api/admin/article-pin`.
5. Reload the current page after success.
6. Write operation log with old and new pin metadata.

Suggested quick actions:

- “全站置顶”: `pinScope=global`, default `pinOrder=100`.
- “分类置顶”: `pinScope=category`, default `pinOrder=100`.
- “取消置顶”: clear all pin metadata.

## Customer-Facing UI

Topic cards should expose pin state through payload fields:

```ts
isPinned: boolean
pinScope?: 'global' | 'category'
```

Suggested display:

- Add a small `置顶` badge before or after the title.
- Use a restrained style, not a full-width banner.
- Keep pinned articles in the same list so pagination behavior stays predictable.

Payload changes:

- Go: `TopicPayload` in `app/http/controllers/forum/payload.go`.
- TS: `TopicPayload` in `resource/src/types/payload.ts`.
- UI: `HomePage.vue`, `CategoryPage.vue`, and any shared topic list rendering.

## Search and Cache Considerations

Pinning is an ordering concern, not a content change.

- Do not rebuild search index only because pinning changed.
- Do invalidate any hot-data cache that feeds home/category article lists.
- If `hotdataserve` keeps cached list data, expose a small invalidation hook or reuse the existing article-list refresh path.

If cached home/category lists are rebuilt periodically, pinning may appear delayed. Admin operations should either invalidate immediately or show this delay clearly.

## Migration Steps

1. Add DB fields and indexes to init/migration SQL.
2. Add fields to `articles.Entity` and `articles.SmallEntity`.
3. Add repository method, for example `UpdatePin`.
4. Update list queries to sort by active pin state first.
5. Add admin endpoint `/api/admin/article-pin`.
6. Add admin runtime API wrapper and types.
7. Add pin controls to post management.
8. Add `isPinned` payload fields and topic-list badges.
9. Add tests for:
   - pinned article appears before normal articles.
   - expired pinned article does not appear as pinned.
   - blocked/draft pinned article still does not appear.
   - category pin only affects categories the article belongs to.
   - non-article-manager cannot pin.

## Open Decision

Choose one before implementation:

1. **Global-only first**: simpler, fastest, enough for homepage operations.
2. **Global + category from day one**: slightly more work, avoids schema/API churn later.

Recommendation: implement **global + category from day one** because the schema cost is small and forum operators usually expect both behaviors.

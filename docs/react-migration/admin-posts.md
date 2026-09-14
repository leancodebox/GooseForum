# Admin posts migration record

## Scope

- React route: `/admin/posts`
- Vue reference: `resource/src/admin/pages/management/PostsManagementPage.vue`

## Preserved behavior

- Topic management and reply review remain separate modes with their original API contracts.
- Both modes preserve moderation-state filtering, search, server paging, page-size selection, refresh, and stale-request protection.
- Topic mode also filters by category and keeps the compact desktop table plus mobile rows.
- Topic actions preserve source viewing/copying, moderation decisions with reasons, one-to-three category assignment, primary-category change confirmation, pin weight, process disable/restore, and confirmed deletion.
- Reply mode preserves content/reason display and approve, reject, or recheck decisions with optimistic controls disabled during writes.

## React boundaries

- Topic, reply-review, source, category, pin, process, delete, and moderation request contracts live in `@gooseforum/client`.
- The admin page is lazy-loaded and preloaded by the SPA sidebar.
- Existing shadcn Dialog, DropdownMenu, Select, Table, Badge, Avatar, Field, and Empty components provide the interaction surfaces without a page-specific theme.

## Verification

- Client tests cover category assignment, pinning, and reply-review route/body contracts.
- No live topic, reply, category, pin, or moderation mutation was performed during migration verification.

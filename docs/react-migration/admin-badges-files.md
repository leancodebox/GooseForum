# Admin badges and file resources migration record

## Scope

- React routes: `/admin/badges`, `/admin/files/resources`
- Vue references: `BadgesManagementPage.vue`, `FileResourcesManagementPage.vue`

## Preserved behavior

- Badges retain system/custom counts, compact honeycomb previews, create/edit fields, automatic/manual grant mode, enabled/wearable flags, sort order, and custom-only deletion.
- File resources retain server pagination, responsive preview cards, image/file rendering, uploader metadata, byte/time formatting, full preview, and absolute URL copying.
- Both pages use the shared admin theme, locale switch, SPA navigation, and lazy route loading.

## Verification boundary

- Badge delete and file pagination contracts are covered by client tests.
- No live badge or file data was modified during migration verification.

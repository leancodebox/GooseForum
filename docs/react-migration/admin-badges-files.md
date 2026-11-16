# Admin badges and file resources migration record

## Scope

- React routes: `/admin/badges`, `/admin/files/resources`
- Vue references: `BadgesManagementPage.vue`, `FileResourcesManagementPage.vue`

## Preserved behavior

- Badges retain system/custom counts, enabled/wearable flags and sort order. The management list includes search, type/status filters and persisted recipient counts; reading the list never aggregates award records.
- The editor offers square 48px built-in icon buttons in an adaptive grid, image upload and an advanced image URL input. The scrollable form reserves space for focus rings. Its card and worn previews share the same appearance component as the public profile; the worn preview uses the default avatar. Codes are generated automatically; codes and sort order appear under advanced settings.
- Custom badges use manual grants. System badges show the fixed automatic condition used by the backend; administrators cannot change that condition through the editor.
- Disabling stops new grants and hides existing profile/worn displays without deleting grant records. Disabling a badge with current recipients requires acknowledgment. Turning off wearing keeps the badge in the member's collection.
- Recipient management supports search, pagination, manual grants and revocation with reasons. It requires both site and user management permissions. Existing user management retains its user-management permission boundary.
- Grant/revoke actions use the existing operation log. The recipient panel lists current grants, not a complete award history. Custom badges with any award history cannot be deleted, including after revocation.
- Badge statistics are stored separately from definition overrides so system badges do not require persisted appearance overrides. Successful grants/restorations increment the stored count, revocations decrement it, and retries do not change it. Award records and statistics change in one transaction; the history flag remains set after revocation. Disabling a badge does not alter its count.
- Existing statistics are initialized once by a versioned migration before startup completes. Request paths must not backfill or recalculate missing statistics. Counts exclude revoked grants and soft-deleted users. There is currently no production account deletion/restoration flow; any future flow must maintain these counters in the same transaction.
- Recipient pagination fetches one extra row to determine `hasNext`. Unfiltered totals come from persisted statistics; filtered searches omit `total` and show no matching count or total page count. Neither path executes a live `COUNT`.
- File resources retain server pagination, responsive preview cards, image/file rendering, uploader metadata, byte/time formatting, full preview, and absolute URL copying.
- Both pages use the shared admin theme, locale switch, SPA navigation, and lazy route loading.

## Verification boundary

- Client tests cover badge recipient pagination without a filtered total and grant/revoke payloads. Backend tests cover validation, persisted counts, deletion protection, idempotent grants, revocation, migration, soft-deleted users and route permissions using SQLite.
- Badge editor and recipient interactions are covered by component tests and desktop/mobile Playwright tests with mocked API responses. Browser checks include creation, uploads, automatic conditions, permissions, disabling acknowledgment and error handling.
- MySQL and PostgreSQL have not been verified against live databases for these changes.
- Browser checks use fixtures and do not grant or revoke real badges or upload live files.

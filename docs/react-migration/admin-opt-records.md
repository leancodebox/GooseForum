# Admin operation records migration record

- React route: `/admin/opt-records`
- Vue reference: `OptRecordsManagementPage.vue`
- Preserved: server pagination, operation/target labels, structured or legacy JSON detail parsing, timestamps, refresh, loading/error/empty states.
- The backend returns a zero-based page in this endpoint; React converts it back to the one-based admin control value.
- Client tests cover the audit-record route and pagination request body.
- No live audit data is mutated by this read-only page.
- Audit localization covers all eight message codes currently emitted by the Go backend plus the legacy moderator status code retained by the Vue implementation. Status values, changed user fields, and Italian target labels are localized; unknown codes fall back to the original log text.

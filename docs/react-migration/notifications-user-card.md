# Notifications and user card migration record

- React component: `notifications.index`; shared component: `UserCardPopover`.
- Vue references: `NotificationsPage.vue`, `UserCardPopover.vue`, `UserCard.vue`, topic avatar stacks, and notification formatting helpers.
- Notifications preserve all/unread filters, unread totals, template-aware copy, per-item and bulk read actions, optimistic rollback, cursor pagination, deduplication, automatic loading, and event-specific targets.
- User cards use the shadcn Popover interaction and open only after a deliberate primary click. Modified clicks preserve native link navigation.
- Card data is fetched only after first opening. Completed responses and concurrent requests are cached by user ID across avatar instances.
- Topic participant avatars and notification actor names now expose the shared card; future topic/post and moderation views can reuse the same component.
- Tests cover individual notification read state, lazy unread-filter loading, user-card fetching, and rendered profile details.

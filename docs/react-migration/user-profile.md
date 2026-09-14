# User profile migration record

- React component: `user.profile`.
- Vue references: `UserPage.vue`, `UserAvatar.vue`, `TopicList.vue`, `TopicListFooter.vue`, badge styling, and social profile icon helpers.
- Preserved: cover and worn badge, identity/status/bio, edit/message/follow actions, website and social links, eight profile statistics, summary topics/badges/activity, full badge directory, and activity tabs for timeline/topics/likes/following/followers.
- Topic activity reuses the shared React topic table. Timeline, topic, and like feeds preserve continuous loading, deduplication, loading feedback, and error states through the framework-neutral page-fetching runtime.
- Internal navigation remains SPA-based, external profile links are restricted to HTTP(S), and follow actions use the shared site API.
- Rendering tests cover summary hierarchy, follow behavior, and topic activity reuse.

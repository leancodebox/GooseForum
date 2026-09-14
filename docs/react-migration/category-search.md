# Category and search page migration record

- React components: `category.index` and `search.index`.
- Vue references: `CategoryPage.vue`, `SearchPage.vue`, `TopicList.vue`, `TopicRow.vue`, `TopicListFooter.vue`, `TopicListModeSwitch.vue`, and `useTopicList.ts`.
- The home, category, and search pages share one React topic-list implementation, including responsive rows, participant avatars, relative activity times, counts, categories, and status indicators.
- Category pages preserve their identity header, category-specific sort labels, hidden category/hot chips, empty state, display-mode preference, pagination, and continuous loading.
- Search preserves result metadata, empty states, standard query URLs, and server pagination. Form submission uses the React runtime so searching does not reload the application shell.
- Rendering tests cover category-specific list behavior and SPA search submission.

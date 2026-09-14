# Home page migration record

- React component: `home.index`
- Vue references: `HomePage.vue`, `TopicList.vue`, `TopicRow.vue`, `TopicListFooter.vue`, `TopicListModeSwitch.vue`, `AvatarStack.vue`, and `useTopicList.ts`
- Preserved: email-verification notice, recent-announcement reminder/read storage, sort tabs, new-topic action, dense responsive topic rows, pin/unseen/hot/category indicators, participant avatars, counts/activity, pagination mode, waterfall mode, automatic loading, deduplication, and stale-response protection.
- The shared React runtime exposes framework-neutral page fetching; Vite supplies it from the development page source and a future Next adapter can provide the same contract.
- A home rendering test covers the core information hierarchy and load-more control.

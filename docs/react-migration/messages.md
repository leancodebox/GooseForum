# Messages page migration record

- React component: `messages.index`.
- Vue reference: `MessagesPage.vue` and its chat API/unread-state helpers.
- The message workspace uses the official shadcn Message and MessageScroller primitives for message semantics, automatic bottom anchoring, preserved scroll position when older messages are prepended, and a scroll-to-latest control.
- Conversation and user results use shadcn Item, scrollable regions use ScrollArea, and search/composer controls use InputGroup.
- Preserved: responsive master/detail layout, conversation search, direct user query entry, suggested-user dialog, initial and older message loading, deduplication, unread clearing, emoji insertion, textarea growth, Enter-to-send, Shift+Enter newline, optimistic local message insertion, and conversation reordering.
- The page is lazy-loaded so the chat-specific primitives do not affect public-page startup.
- Tests cover message loading, unread acknowledgement, Enter-to-send, and starting a new conversation.

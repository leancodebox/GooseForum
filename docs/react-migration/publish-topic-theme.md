# React publish, topic, and theme migration

## Publish and editor

- Supports create/edit payloads, normal and restricted category selection, main-category confirmation, validation, draft saving, publishing, and moderation responses.
- The shared composer offers Markdown and ProseMirror visual modes, Markdown preview, formatting shortcuts/toolbars, tables, rich clipboard conversion, image paste/drop/multi-upload, client-side image optimization, and upload progress.
- SPA navigation and browser unload are guarded while an editor has unsaved changes. The user can continue editing, discard, or save a draft before leaving.

## Topic detail

- Preserves bidirectional post-window loading, automatic forward loading, scroll-position-aware active post tracking, and a shadcn Slider for jumping to a post window.
- Supports reply targets, creating/editing/deleting replies, topic editing/deletion, optimistic like/bookmark/watch actions, user cards, reports, moderator block/restore actions, server-rendered Markdown, Mermaid enhancement, and image preview navigation.
- The reply composer is a responsive floating dock with minimized and expanded states and reuses the publish editor/upload behavior.

## Theme preview

- Preserves all nine existing presets, light/dark editing, enable state, every `SiteThemeTokenKey`, radius/depth controls, contrast checks, forum/component/CSS previews, restore/default actions, draft saving, and site publishing.
- Preview values use the existing `--gf-*` theme protocol and are removed when leaving the page. No alternate React-only theme format was introduced.

## Intentional implementation differences

- React 19 state and host runtime hooks replace Vue watchers and directives.
- shadcn Dialog, Tabs, ToggleGroup, Slider, Field, Alert, and other official primitives provide focus and keyboard behavior while the GooseForum theme tokens retain product styling.

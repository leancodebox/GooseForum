# Admin links and sponsors migration record

## Scope

- React routes: `/admin/links`, `/admin/sponsors`
- Vue references:
  - `resource/src/admin/pages/management/LinksManagementPage.vue`
  - `resource/src/admin/pages/management/SponsorsManagementPage.vue`

## Preserved behavior

- Link groups support create, edit, delete, ordering, emoji/color selection, and entry counts.
- Links support create, edit, delete, visible/hidden state, ordering within a group, and movement between groups. Link changes persist immediately and reload server state after a failed save.
- Sponsor-page title, description, four tiers, contact section, button content, and rules remain editable.
- Sponsors support create, edit, remove, image URL/upload, ordering within a tier, and movement between tiers. Sponsor edits remain local until **Save all**, matching the original page.
- Public destination links continue opening in a separate tab.

## Dependencies and boundaries

- Drag-and-drop uses the React application's existing `@dnd-kit/core`, `@dnd-kit/sortable`, and `@dnd-kit/utilities` dependencies; no second drag library was introduced.
- The shared React package adds the official shadcn Textarea source component without an additional runtime dependency.
- Link, sponsor, and admin-image-upload contracts live in the framework-neutral `@gooseforum/client` admin API.
- Both pages are lazy-loaded and preloaded by the existing SPA sidebar.

## C-side visual alignment

- Admin link previews reuse the C-side `2 / 3 / 4 / 5` responsive grid, 32px logo, `rounded-xl`, `p-2`, title/description typography, border, and hover treatment.
- Sponsor previews reuse the C-side tier-specific grids, card padding, and 44px / 40px / 32px avatar scale.
- Admin-only drag/edit/delete controls are hover/focus overlays, so they do not consume preview-card content width or change the public card height.
- Supporting panels use the same 260px column, `rounded-xl`, and `p-4` surface treatment as the public pages.

## Verification

- Client tests cover link/sponsor request envelopes and multipart image upload headers.
- No live link, sponsor, ordering, or image configuration was changed during migration verification.

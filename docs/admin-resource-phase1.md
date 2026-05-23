# Admin Resource Phase 1

## Goal

Move toward a Vue-based admin surface inside the `resource` frontend system while keeping the current React admin at `/admin/` available during migration.

Phase 1 is intentionally small:

- Make the current customer-facing frontend explicit as the `site` entry.
- Add an independent Vue admin entry that does not load for normal customer pages.
- Ship the first admin layout page as a new parallel route.
- Keep shared code opt-in. Only code imported by both entries should live in `shared`.

## Non-goals

- Do not remove or rewrite the existing `admin/` React application.
- Do not point `/admin/` to the new Vue admin yet.
- Do not merge site and admin into one browser entry.
- Do not redesign admin feature workflows in this phase.

## Target Structure

```text
resource/src/
  site/
    main.ts
    App.vue
    pages/
    components/
  admin/
    main.ts
    AdminApp.vue
    pages/
    components/
  shared/
    api/
    format/
    ui/
  runtime/
  styles/
  types/
```

`runtime`, `styles`, and `types` may stay in place during phase 1. They can move to `shared` later when ownership is clearer.

## Routing

- Existing React admin remains at `/admin/`.
- New Vue admin starts at `/manage`.
- `/manage` uses a separate Go template and loads only `src/admin/main.ts`.
- Customer-facing pages load only `src/site/main.ts`.

## Acceptance Criteria

- `pnpm --dir resource build` passes.
- `pnpm --dir resource typecheck` passes.
- Existing site pages still resolve through the site entry.
- `/manage` renders the first Vue admin layout page for authenticated admins.
- `/admin/` continues to serve the existing React admin.

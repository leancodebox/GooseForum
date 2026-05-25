# Resource Admin Current State

## Status

The admin surface is a first-class Vue admin inside `resource`.

## Entries

- Customer-facing pages use `resource/src/site/main.ts`.
- Admin pages use `resource/src/admin/main.ts`.
- The admin template is `resource/templates/pages/admin.gohtml`, wrapped by `resource/templates/layout/admin.gohtml`.
- The admin template loads only the admin entry through `ResourceEntry "src/admin/main.ts"`.

## Routing

- `/admin` and `/admin/*path` render the resource admin shell.
- Admin routing is handled client-side by `resource/src/admin/runtime/router.ts`.
- The Go entry point is `forum.Manage` in `app/http/controllers/forum/manage.go`.
- The route is protected by login and admin permission checks in `app/http/routes/route4api.go`.

## Structure

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
    layouts/
    components/
    pages/
    runtime/
    styles/
    types.ts
  runtime/
  styles/
  types/
```

`resource/src/runtime`, `resource/src/styles`, and `resource/src/types` are shared by the resource app and should stay framework-neutral. Admin-only code belongs under `resource/src/admin`.

## API Ownership

- Public/customer APIs live under `/api/forum` or top-level `/api/*` routes.
- Admin APIs live under `/api/admin`.
- Admin API wrappers live in `resource/src/admin/runtime/api.ts`.
- Customer-facing API wrappers live in `resource/src/runtime/api.ts`.

## Implementation Rules

- Do not introduce migration-era names, fields, route prefixes, or labels. The resource admin is the canonical admin path now.
- Keep customer and admin browser entries separate. Admin-only dependencies should not be imported from `resource/src/site`.
- Keep shared code opt-in. Only code imported by both entries should live outside `resource/src/admin` and `resource/src/site`.
- For admin pages, maintain route metadata and navigation in one place before considering a page complete.
- For management pages, prefer focused admin API endpoints plus typed wrappers in `resource/src/admin/runtime/api.ts`.
- After changing `resource/`, run `pnpm --dir resource typecheck` and `pnpm --dir resource build` when dependency state allows it.

## Current Known Gap

At the time this document was updated, `resource/src/admin` references UI/runtime dependencies that are not installed in `resource/package.json`, including packages such as `reka-ui`, `@vueuse/core`, `vuedraggable`, `vue-sonner`, `@internationalized/date`, `@unovis/vue`, `clsx`, `tailwind-merge`, and `class-variance-authority`.

Until those dependencies are installed or the references are removed, full `pnpm typecheck` / `pnpm build` for `resource` will fail before reaching unrelated site changes.

# Resource Frontend Design

This document describes the current frontend architecture of GooseForum. The `resource/` application is the canonical frontend for both the public site and the admin console.

## Goals

- Keep GooseForum deployable as a single Go binary with embedded frontend assets.
- Use Vue 3 for interactive pages while keeping GoHTML output available for SEO and no-js fallback.
- Keep the public site and admin console as separate entry points so each side only loads the code and styles it needs.
- Share stable runtime utilities, payload types, API helpers, and low-level primitives where they are genuinely common.
- Avoid coupling public-site styling to admin styling. Shared code should be extracted deliberately, not by accident.

## Non-Goals

- Do not introduce an independent Node SSR server.
- Do not route the admin console through a separate React app.
- Do not use migration-era route prefixes or compatibility names in new code.
- Do not make every visual pattern a shared component. Components are extracted when reuse, consistency, or complexity justifies it.

## Current Structure

```text
resource/
  src/
    site/                 Public Vue entry, pages, and site-only components
    admin/                Admin Vue entry, pages, layout, styles, and components
    runtime/              Shared payload runtime and browser helpers
    styles/               Public-site styles
    types/                Shared payload and entity types
    locales/              Frontend locale resources
  templates/
    layout/app.gohtml     Public-site shell
    layout/admin.gohtml   Admin shell
    pages/*.gohtml        No-js and initial HTML pages
    partials/*.gohtml     Template fragments
```

The public site is mounted from `resource/src/site/main.ts`.

The admin console is mounted from `resource/src/admin/main.ts` and is served under `/admin` and `/admin/*path`.

## Rendering Model

Go controllers render GoHTML for the initial document. When JavaScript is enabled, the Vue app reads the embedded page payload and takes over navigation.

For in-app navigation, the frontend requests the same route with the `X-Goose-Page` header. The backend returns a JSON payload instead of the full HTML document. The runtime applies the new payload without rebuilding the whole shell.

This gives GooseForum three useful modes:

- Normal browser load: full HTML document.
- SPA navigation: JSON payload with smooth in-app transitions.
- No-js/crawler access: readable GoHTML output.

## Payload Contract

Page payloads should be treated as the contract between Go controllers and Vue pages.

Rules:

- Keep payloads page-focused and serializable.
- Prefer business data and navigation state over pre-rendered UI strings.
- Keep layout payload data stable: current user, navigation, settings, permissions, and shared shell metadata.
- Do not add frontend-only temporary flags to backend payloads unless they describe real server state.
- When a payload field is removed or renamed, update Go structs, Vue types, templates, and docs in the same change.

## Routing

Public routes are rendered by Go controllers and enhanced by the site SPA runtime.

Admin routes are protected by login and admin permission middleware, then rendered by the admin shell. The admin frontend owns the internal page switch after the shell is loaded.

Routing rules:

- Public paths should remain meaningful URLs.
- Admin paths should stay under `/admin`.
- Route labels shown in the UI should come from the current route/menu metadata, not from raw path segments.
- Unknown admin subpaths should show an admin-side not-found state or redirect to the admin dashboard.

## Public Site

The public site should prioritize reading, scanning, and participation.

Key areas:

- Home and category topic lists.
- Topic detail and post stream.
- Publish/edit flows.
- Search.
- User profile and settings.
- Messages and notifications.

Public-site implementation rules:

- Keep no-js output useful for core content.
- Keep page titles, canonical metadata, and topic content available in GoHTML.
- Use the shared payload runtime for navigation and API error handling.
- Keep public-site CSS in `resource/src/styles` and site-scoped components.

## Admin Console

The admin console is an operational tool. It should be dense, predictable, and stable under repeated use.

Key areas:

- Dashboard and site statistics.
- Topic management.
- Category management.
- User management.
- Settings and operational configuration.

Admin implementation rules:

- Keep admin CSS in `resource/src/admin/styles`.
- Keep admin layout, sidebar, header, and shell state under `resource/src/admin`.
- Use shared utilities only when the behavior is truly shared with the public site.
- Prefer established libraries for complex widgets such as charts, date selection, drag sorting, and accessible overlays.
- Wrap third-party widgets behind local components when GooseForum needs consistent styling or a stable internal API.

## Component Policy

Extract a component when at least one of these is true:

- The same behavior appears in multiple places.
- The behavior has several interaction states.
- Accessibility or keyboard handling would otherwise be duplicated.
- A third-party library needs a local wrapper.
- The component represents a stable product concept such as topic row, admin page header, or category selector.

Do not extract a component only because a block of markup is visually large. A one-off page section can stay local if it has no shared behavior.

## Styling Boundaries

Public-site and admin styling are intentionally separate.

Rules:

- Public-site global styles live in `resource/src/styles`.
- Admin global styles live in `resource/src/admin/styles`.
- Shared design tokens may be introduced only when both sides intentionally use the same semantic value.
- Do not rely on public-site selectors to style admin pages.
- Do not rely on admin selectors to style public pages.

## Build and Development

Install and run the resource frontend from `resource/`:

```bash
cd resource
pnpm install
pnpm dev
```

Build production frontend assets:

```bash
cd resource
pnpm build
```

Build the Go binary from the repository root:

```bash
go build -ldflags="-w -s" .
```

When `[app].env = "local"` in `config.toml`, the Go templates should load Vite development assets. In production, run the resource build before building or serving the final binary.

## Maintenance Rules

- Documentation follows the current code, not migration plans.
- Remove obsolete route names, temporary prefixes, and retired stack references when the code changes.
- Update README files, architecture docs, and UI specs together when frontend ownership changes.
- Keep docs explicit about known gaps when implementation and intended behavior differ.

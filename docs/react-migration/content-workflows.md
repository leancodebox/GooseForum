# React content workflow migration

## Covered pages

- `drafts.index`: compact draft table, status/category metadata, empty state, and edit/new-draft navigation.
- `access-groups.index`: parallel joinable/managed group loading, applications, manager approval/rejection, and error/loading states.
- `moderation.index`: reports, open/closed filters, target blocking, report rejection, blocked-topic restoration, audit log pagination, and guidance.
- `error.index`: localized titles and server message codes with back/home recovery actions.

## Compatibility notes

- The page payload and site API contracts remain in `@gooseforum/client`; the React components do not depend on a Vite or Next host.
- Existing Vue routes and templates are unchanged.
- The four existing locales are sourced from the established Vue copy so the React pages keep the current terminology.

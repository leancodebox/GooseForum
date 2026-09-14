# Admin users migration record

## Scope

- React route: `/admin/users`
- Vue reference: `resource/src/admin/pages/management/UsersManagementPage.vue`

## Preserved behavior

- User listing remains server-paginated with username search, clear, refresh, page-size selection, range display, and previous/next controls.
- Desktop keeps the compact user, role, status, creation time, last-active time, and action columns. Mobile uses the original condensed user rows.
- User editing preserves account enabled/disabled state, email verification, role assignment, account metadata, prestige, automatic badges, and selectable manual badges.
- Role options and badge options load independently so one optional panel failure does not prevent editing the other user fields.
- User fields are saved first; manual badges are saved only after badge options loaded successfully, matching the previous failure boundary.

## React boundaries

- User list/edit/role-option/badge endpoints live in the framework-neutral `@gooseforum/client` admin API.
- The page is lazy-loaded and preloaded from the SPA sidebar on hover or keyboard focus.
- The editor uses shared shadcn Dialog, Switch, Select, Avatar, Badge, and Empty components with the existing site theme tokens.

## Verification

- Client tests cover user edit, badge-option, and manual-badge request contracts.
- No live user or badge data was changed during migration verification.

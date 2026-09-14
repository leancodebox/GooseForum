# Admin access groups migration record

## Scope

- React route: `/admin/access-groups`
- Vue reference: `resource/src/admin/pages/management/AccessGroupsManagementPage.vue`

## Preserved behavior

- The page keeps the original two-column desktop structure: searchable access-group navigation on the left and the selected group detail on the right.
- System groups remain immutable and expose only their category-permission summary.
- Custom groups support create, edit, enable/disable, and confirmed deletion with the original `invite_only` and `application` join modes.
- The detail area keeps separate category-permission and member tabs. Category permissions are read-only here and link to `/admin/categories` for editing.
- Members are ordered by status, role, and user ID. Administrators can add a username or numeric user ID, choose member/manager role, remove active members, and approve or reject pending applications.
- Data refreshes preserve the selected group when it still exists and fall back to the first available group after deletion.

## React boundaries

- Access-group routes and request bodies are typed in the framework-neutral `@gooseforum/client` admin API.
- Page composition and interaction state stay in the React Vite admin application.
- The route is lazy-loaded and participates in the existing History API navigation and hover/focus preloading.

## Theme and components

- All surfaces use the existing payload theme and semantic shadcn tokens.
- The shared React package now includes the official shadcn `Dialog` and `Switch` source components required by the original interaction model.
- No independent access-group palette or page-level theme was introduced.

## Verification

- `@gooseforum/client`: 34 tests passed, including access-group save/member/review request contracts.
- React admin routing: 5 tests passed.
- Client and React TypeScript checks passed.

Mutation flows were validated through typed request contracts only; no running access-group or member data was changed.

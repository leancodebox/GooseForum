# Admin categories migration record

## Scope

- React route: `/admin/categories`
- Vue references:
  - `resource/src/admin/pages/management/CategoriesManagementPage.vue`
  - `resource/src/admin/components/CategoryAccessDialog.vue`
  - `resource/src/admin/components/layout/AppSidebar.vue`
  - `resource/src/admin/runtime/access.ts`

## Preserved behavior

- Navigation and route visibility follow the six existing admin permission values; administrator permission remains a superuser permission.
- A topics manager can edit categories and manage global/category moderators.
- A role manager can inspect and update category access without receiving category-edit privileges.
- Category access continues to derive public/restricted state from the `everyone` group, excludes disabled groups from writes, warns when no enabled group can read, and blocks a public-to-restricted change while multi-category topics still exist.
- Moderator user lookup accepts either a username or a numeric user ID and remains debounced.
- Unauthenticated users are redirected to login with the original admin URL; authenticated users without a route permission are redirected to their first permitted admin page.

## React boundaries

- Admin HTTP contracts live in the framework-neutral `@gooseforum/client` package.
- The admin page and its routing shell remain in the React Vite application because they are application-specific composition, not reusable UI primitives.
- Dashboard and category page code are loaded as separate chunks so category management does not download the chart/table demo bundle.

## Theme and localization

- The existing `admin.shell` payload still drives `data-theme`, the optional `/site-theme.css`, and all semantic shadcn color tokens.
- No independent admin palette was introduced. A category's own color is only used for its category marker and color input.
- The new shell/category strings select Chinese, English, Japanese, or Italian from the existing browser/cookie locale convention; Japanese and Italian currently fall back to English for untranslated secondary descriptions.

## Follow-up layout polish

- Sidebar navigation uses a balanced 30px item rhythm with a 2px inter-item gap and slightly relaxed group spacing, between the first React draft and the older Vue layout.
- The category page now uses one flat management surface: its title, count, search, and primary actions share a compact toolbar instead of being split across a page header and a nested card header.
- The desktop category table follows the original compact information architecture: ID, category name, visibility, slug, one-line description, moderators, sort, and actions each have a stable column. Rows and headers use the shorter management-table rhythm, while access/edit/delete actions remain available without expanding the row.
- Below the `md` breakpoint, categories render as flat divided rows with wrapped action buttons instead of nested cards or a horizontally scrolling table.
- Empty search results use the shared shadcn `Empty` composition.
- Ordinary admin menu clicks use History API navigation and update only the content region. Modified clicks retain native link behavior, browser back/forward is supported, and migrated lazy routes preload on hover or keyboard focus.

## Verification

- `@gooseforum/client`: 33 tests passed, including admin route and request-body contracts.
- `@gooseforum/react`: 21 tests passed.
- React admin permission routing: 3 tests passed.
- Client and React application TypeScript checks passed.
- React Vite production build passed.
- Logged-out `/admin/categories` smoke check redirected to `/login?redirect=%2Fadmin%2Fcategories`.

Mutation flows were verified with mocked request contracts only; no category, moderator, or access data was changed in the running service.

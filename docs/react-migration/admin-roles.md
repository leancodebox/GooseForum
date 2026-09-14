# Admin roles migration record

## Scope

- React route: `/admin/roles`
- Vue reference: `resource/src/admin/pages/management/RolesManagementPage.vue`

## Preserved behavior

- Role and permission-option requests start in parallel.
- Name search, enabled/disabled filtering, client-side pagination, page-size selection, and manual refresh remain available.
- The table keeps ID, role name, status, permission badges, creation time, and edit/delete actions.
- Create/edit validates a non-empty role name and at least one permission.
- Delete remains a confirmed destructive action.

## Shared admin shell

- The React top bar now restores the Vue shell's light/dark theme toggle, four-language selector, and desktop return-to-site action.
- Theme changes reuse the site theme cookie, local storage key, payload colors, and `data-theme` contract.
- Locale changes update the cookie, `html.lang`, header labels, navigation, and migrated page strings without a full reload.
- The header separator explicitly overrides the shared vertical separator's stretch behavior so its 16px line is centered.

## Verification

- Role list, permission list, save, and delete request contracts are covered by the framework-neutral client tests.
- No live role mutation was performed during migration verification.

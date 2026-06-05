# Resource UI Specification

This document defines the current executable UI rules for GooseForum `resource/`. It covers both the public site and the admin console.

If implementation and this document disagree, verify the current product intent first, then update the code and this document together.

## Product Feel

GooseForum should feel:

- Content-first.
- Quiet and readable.
- Dense enough for daily use.
- Fast to scan.
- Stable across navigation and data refreshes.

Avoid:

- Marketing-style hero sections for application pages.
- Decorative backgrounds that compete with content.
- Card-heavy layouts where rows or sections would be clearer.
- Page-specific one-off styling that cannot be maintained.
- Large title blocks inside operational admin screens.

## Frontend Areas

The public site and admin console share the same repository and build system, but their UI boundaries are separate.

```text
resource/src/site/          Public pages and public-only components
resource/src/styles/        Public global styles
resource/src/admin/         Admin pages, layout, components, and styles
resource/src/runtime/       Shared payload/runtime helpers
resource/src/types/         Shared payload and entity types
```

Shared UI code is allowed only when it has a stable purpose on both sides. Admin layout details should not leak into the public site.

## Design Tokens

Use Tailwind's default breakpoints unless project configuration explicitly changes them:

```text
sm: 640px
md: 768px
lg: 1024px
xl: 1280px
2xl: 1536px
```

Spacing:

```text
xs: 4px
sm: 8px
md: 12px
lg: 16px
xl: 24px
2xl: 32px
3xl: 48px
```

Radius:

```text
Small controls: 6px
Buttons and inputs: 8px
Panels: 8px to 12px
Avatars: full
```

Typography:

```text
Body: 14px or 15px
Metadata: 12px to 13px
Topic title: 15px mobile, 16px desktop
Page title: 24px mobile, 28px to 30px desktop
Admin page title: compact, usually 20px to 24px
```

Rules:

- Font size must not scale with viewport width.
- Letter spacing should stay normal unless a small label truly needs it.
- Numbers in statistics and tables should use tabular numerals where practical.
- Article content uses prose styles, not generic page text styles.

## Color and Surfaces

Public site:

- Background: quiet neutral.
- Content surface: white or near-white.
- Text: strong neutral for primary content, muted neutral for metadata.
- Accent: blue for primary actions and active navigation.
- Category colors: small identifying marks only.

Admin console:

- Background and shell colors come from admin scoped CSS variables.
- Sidebar, header, panels, and overlays use admin-local tokens.
- Operational states should be obvious: success, warning, danger, disabled, loading.

Rules:

- Use borders before shadows for structure.
- Shadows should be restrained and reserved for overlays or focused floating UI.
- Do not use large gradients, decorative blobs, or purely atmospheric visuals.

## Motion

Motion should make state changes legible without making the app feel soft or delayed.

Rules:

- Maintain motion rules globally, but implement motion at the component that owns the state.
- `resource/src/runtime/motion.ts` is the source of truth for public-site timing, easing, and shared motion presets.
- App shell and route-level motion should only handle navigation and shell stability. Local components should own local transitions such as menus, drawers, reply editors, hover cards, and flash messages.
- Avoid a centralized animation controller that reaches into component internals.
- Avoid page-specific one-off transitions unless the page has a unique interaction model.
- Layout transitions should animate width, transform, opacity, or grid tracks predictably.
- Sidebar collapse must not make icons jump vertically or horizontally.
- Page switches should avoid full white flashes.
- Data refresh should preserve the old content until the new state is ready when practical.
- Respect `prefers-reduced-motion`.

Recommended durations:

```text
Fast control feedback: 120ms to 160ms
Content enter and small expansion: 140ms to 180ms
Panel/sidebar transitions: 180ms to 240ms
Overlay enter/leave: 160ms to 220ms
Chart or metric update: 200ms to 320ms
```

Public-site motion layering:

```text
runtime/motion.ts
  Defines shared durations, easing, and motion presets.

AppShell / PayloadRouteView / navigation-state
  Handles route transition, loading visibility, and shell stability.

Local site components
  Use shared motion tokens while owning their own DOM transitions.
```

Public SPA navigation rules:

- Keep the app shell mounted across in-app navigation.
- Do not clear shell header state before the next page is ready to render.
- Animate the main content area only when the page identity changes, such as topic list to article detail.
- Do not animate the whole page for same-view filter changes. Stable regions such as announcements, tabs, headers, and list frames should stay still when only list data changes.
- Let the component that owns the changed data provide local motion when it does not disturb document height. For same-view topic sort changes, do not animate the whole topic list in or out; prefer direct data replacement plus stable tab/loading feedback.
- Keep the previous page visible until the next payload and component are ready.
- Use subtle opacity and small vertical movement for page content. Avoid scale, large movement, or whole-shell animation.
- Sidebar active state may transition color/background, but must not change item size or position.
- Loading indicators should appear only after a short delay and remain visible long enough to avoid flicker.

## Public App Shell

Desktop:

```text
header
main
  optional left navigation
  page content
  optional right rail
```

Mobile:

```text
header
main
drawer or compact navigation
```

Rules:

- Header remains stable during SPA navigation.
- Main content should keep readable width on large screens.
- Right rail appears only when it provides useful context.
- Mobile controls must not overflow their containers.

## Admin App Shell

Admin layout is a fixed operational shell:

```text
admin shell
  sidebar
  main column
    header
    scrollable content
```

Rules:

- Sidebar and header stay fixed relative to the shell.
- Only the content area scrolls.
- Collapse and expand behavior must keep logo, icons, and active states aligned.
- Mobile admin navigation uses an overlay or drawer pattern.
- Route title/breadcrumb should use menu metadata and Chinese labels, not raw path text.

## Topic Lists

Topic list rows must support:

- Title.
- Excerpt or latest reply snippet.
- Category.
- Tags when available.
- Reply count.
- View count.
- Last activity time.
- Participant avatars when available.

Rules:

- Desktop lists may use table-like columns.
- Mobile lists merge statistics into compact metadata.
- Rows should be dense and scannable, not oversized cards.
- Hover and active states should be subtle.
- Empty states explain the reason and offer a useful action when possible.

## Article Detail

Article detail structure:

```text
title block
article body
reply stream
reply composer or login prompt
optional right rail
```

Rules:

- The title is prominent but not oversized.
- Article body and editor preview must share prose behavior.
- Reply actions should be available without visually dominating the stream.
- A user's own replies may expose edit actions when permissions allow.
- Errors from reply creation or editing must be shown near the action that failed.

## Forms

All forms must support:

- Label.
- Help text when needed.
- Field-level error text.
- Focus-visible state.
- Disabled state.
- Pending state.
- Success or saved feedback where appropriate.

Rules:

- Do not rely on placeholder text as the only label.
- Prevent duplicate submission during pending state.
- Client validation should catch obvious mistakes before the request.
- Server validation messages should be surfaced instead of swallowed.

## Admin Data Views

Tables and data-heavy pages should use reliable libraries or local wrappers around reliable libraries when the behavior is complex.

Rules:

- Use real table semantics for tabular data.
- Provide loading, empty, error, and pagination states.
- Keep filters close to the data they affect.
- Date range inputs must prevent invalid ranges.
- Chart widgets should use a maintained charting library wrapped behind local admin components.

## Component Extraction

Extract a component when:

- It appears in more than one page.
- It owns keyboard, focus, overlay, or transition behavior.
- It wraps a third-party library.
- It represents a stable product concept.
- It reduces repeated API-state handling.

Keep code local when:

- The UI is page-specific.
- The markup is simple.
- Extraction would create a vague component with too many props.

## Accessibility

Requirements:

- Use semantic headings.
- Use real buttons for actions and links for navigation.
- Menus, drawers, dialogs, and popovers must be keyboard operable.
- Icon-only buttons must have accessible names.
- Focus state must be visible.
- Color contrast must remain readable.
- Critical actions cannot be available only on hover.

## Internationalization

Rules:

- Do not hardcode English labels in reusable components.
- Backend payloads may provide business names and user-generated content.
- Frontend dictionaries or route/menu metadata should provide common UI labels.
- Long Chinese, English, and Japanese strings must be considered in buttons, tabs, and sidebars.

## Page Acceptance

Public topic pages:

- Desktop and mobile layouts preserve title, category, stats, and activity.
- No-js output contains the core list.
- SPA navigation does not flash blank content.

Article pages:

- Article content and replies are readable without JavaScript.
- Reply creation and own-reply editing surface validation errors.
- Composer pending and failure states are clear.

Admin pages:

- Sidebar/header remain stable while content scrolls.
- Current route label is localized.
- Filters, tables, charts, and dialogs have loading/empty/error states.
- Complex widgets are wrapped or isolated so future replacement is possible.

## Maintenance Rules

- Keep public and admin CSS separate unless a shared token is deliberate.
- Update this document when a reusable component pattern changes.
- Remove obsolete migration wording when code becomes canonical.
- Prefer current implementation truth over stale design notes.

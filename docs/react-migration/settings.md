# User settings migration record

- React component: `settings.index`.
- Vue references: `SettingsPage.vue`, `useAvatarCropUpload.ts`, user avatar/badge helpers, and the existing settings API methods.
- Preserved: profile identity preview, cover editing, custom and preset avatars, wearable badge selection, statistics, username/email editing and verification, public profile fields, locale and social links, password change/reset, local privacy preferences, OAuth bindings, and authorized OIDC applications.
- Avatar cropping continues to use the established CropperJS workflow and uploads 300 px and 96 px variants through the shared client API.
- Settings tabs update the URL without replacing the page payload. OAuth bindings and OIDC grants load together only after their settings surface is opened.
- Forms use shared shadcn Field controls; status/error feedback, destructive actions, overlays, and empty/loading states use the corresponding shared primitives.
- Tests cover profile saving, deferred connection loading, OAuth disconnect, and authorized application rendering.

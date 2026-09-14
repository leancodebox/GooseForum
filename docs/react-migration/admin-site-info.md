# Admin site information migration record

- React route: `/admin/settings/site-info`
- Vue reference: the `site-info` branch in `AdminSettingsPage.vue`
- Preserved fields: site name, public URL, contact email, logo URL/upload, description, keywords, and external-links JSON.
- Settings and upload contracts live in `@gooseforum/client`; the page uses the shared admin theme, locale selector, and SPA routing.
- No live site settings were changed during migration verification.

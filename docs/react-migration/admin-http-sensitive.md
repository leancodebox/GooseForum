# Admin HTTP notifications and sensitive words migration record

- React routes: `/admin/settings/http-notify`, `/admin/settings/sensitive-words`
- Vue references: HTTP-notify branch in `AdminSettingsPage.vue` and `SensitiveWordsSettingsPage.vue`
- HTTP notifications preserve endpoint enablement, URL/secret/events/timeout fields, validation, failure state display, normalization, and configuration/guide tabs.
- Sensitive words preserve async review enablement/mode, search, 50-item paging, reject/replace/record actions, replacement text, inline editing, enablement, and deletion.
- Locale changes do not reload unsaved forms. Client tests cover both settings and word mutation envelopes; no live configuration was changed.

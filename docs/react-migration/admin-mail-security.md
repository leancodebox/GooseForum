# Admin mail and security migration record

- React routes: `/admin/settings/mail`, `/admin/settings/security`
- Vue reference: corresponding branches in `AdminSettingsPage.vue`
- Mail preserves enablement, SMTP host/port/credentials, SSL, sender identity, and testing with current unsaved form values.
- Security preserves signup, required email verification, and normalized unique email-domain allowlisting.
- Locale changes update labels without reloading or discarding unsaved settings.
- Client tests cover test-mail and security-save request bodies. No live settings were changed during verification.

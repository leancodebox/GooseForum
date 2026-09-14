# Admin posting and announcement migration record

- React routes: `/admin/settings/posting`, `/admin/settings/announcement`
- Vue reference: corresponding branches in `AdminSettingsPage.vue`
- Posting preserves title/body limits, new-user cooldown, daily topic limit, attachment enablement, upload limits, file size, cooldown, and normalized extension allowlisting.
- Announcement preserves enablement, content editing, and example insertion.
- Locale changes do not reload or discard unsaved forms. Client tests cover both save envelopes; no live settings were changed.

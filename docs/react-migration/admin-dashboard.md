# Admin dashboard migration record

- React route: `/admin`
- Vue references: `StatsPage.vue`, `TrafficOverview.vue`, `ProjectVersion.vue`
- Replaces the shadcn demo with real users/topics/replies/links statistics, monthly deltas, date-bounded daily traffic, server version/mode/commit, and GitHub releases.
- Statistics, traffic, version, and releases load independently so an external release failure does not hide local metrics.
- Dashboard contracts live in `@gooseforum/client`; client tests cover local statistics, traffic, and version routes.
- Traffic uses the shared shadcn Chart wrapper with themed series, accessible Recharts output, localized axes/tooltips, and range totals for registrations, topics, and replies.
- Date filtering uses shadcn Calendar + Popover with an inclusive range picker and 7/30/90-day shortcuts; custom ranges are applied explicitly to avoid duplicate API requests while selecting.

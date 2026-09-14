# Admin dashboard migration record

- React route: `/admin`
- Vue references: `StatsPage.vue`, `TrafficOverview.vue`, `ProjectVersion.vue`
- Replaces the shadcn demo with real users/topics/replies/links statistics, monthly deltas, date-bounded daily traffic, server version/mode/commit, and GitHub releases.
- Statistics, traffic, version, and releases load independently so an external release failure does not hide local metrics.
- Dashboard contracts live in `@gooseforum/client`; client tests cover local statistics, traffic, and version routes.

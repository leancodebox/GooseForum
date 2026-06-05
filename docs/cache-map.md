# Cache Map

This document records the current in-process and response-cache assets. It is
intended to keep cache ownership, stale budgets, and invalidation paths explicit.

## Cache Layers

| Layer | Implementation | Main use | Notes |
| --- | --- | --- | --- |
| Typed data cache | `app/bundles/localcache` | Config, permissions, lists, SEO XML, short-lived UI state | Backed by `ttlcache/v3`, typed values, per-cache capacity, per-call TTL, singleflight. |
| Static response cache | `app/http/middleware/gzipcache.go` | Production embedded static responses | Stores gzip responses in memory by method and path. Embed content is finite, so it has no cleanup. |
| Browser cache headers | `BrowserCache`, `httputil.SetLongPublic` | Static assets and existing file images | Sets long public `Cache-Control`; this is not an application memory cache. |

## Business Caches

| Cache | Owner | Value | TTL | Capacity | Key space | Source | Invalidation | Stale policy |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Site settings config | `hotdataserve` | `pageConfig.SiteSettingsConfig` | 5s | 4 | Finite: one key | `page_config` | Admin save clears | Should hot update quickly. |
| Mail settings config | `hotdataserve` | `pageConfig.MailSettingsConfig` | 5s | 4 | Finite: one key | `page_config` | Admin save clears | Should hot update quickly. |
| Announcement config | `hotdataserve` | Prepared announcement config | 5s | 4 | Finite: one key | `page_config` | Admin save clears | Prepared HTML is cached with config. |
| Security settings config | `hotdataserve` | Registration/security config | 5s | 4 | Finite: one key | `page_config` | Admin save clears | Short TTL keeps permission checks fresh. |
| Posting settings config | `hotdataserve` | Posting/content config | 5s | 4 | Finite: one key | `page_config` | Admin save clears | Short TTL keeps posting checks fresh. |
| Sponsors config | `hotdataserve` | Sponsors config | 1m | 4 | Finite: one key | `page_config` | Admin save clears | Presentation data; minute-level stale is fine. |
| Friend links config | `hotdataserve` | Friend link groups | 1m | 4 | Finite: one key | `page_config` | Admin save clears | Also feeds site statistics link count. |
| Site statistics | `hotdataserve` | `vo.SiteStats` | 5s | 4 | Finite: one key | Users, articles, replies, daily stats, friend links | TTL only | Short snapshot for homepage/sidebar display. |
| Article list VO | `hotdataserve` | `[]*vo.ArticlesSimpleVo` | 5s | 512 | Capped input: sort plus page, only first 50 pages cached | Article page query, users, categories | Pin/category changes clear; other article writes rely on TTL | Deliberately a short-lived presentation snapshot. Do not increase TTL without adding dependency invalidation. |
| Article categories | `hotdataserve` | Category snapshot with list and map accessors | 1m | 4 | Near-finite: one snapshot | `article_category` | Admin category save/delete clears | Small data; one cache entry serves both list and map callers. |
| SEO XML | `seoController` | sitemap/rss XML string | 10s | 128 | Input-derived: host, guarded by host validation and capacity | Articles, categories, site settings | TTL only | SEO output may briefly lag behind writes. |
| Role permissions | `permission` | `[]permission.Enum` by role id | 10m | 256 | DB-bounded: role id | `role_permission_rs` | Role save/delete and admin role setup delete role key | Correctness-sensitive; explicit invalidation is required on role edits. |
| User info | `userservice` | `userservice.UserInfo` | 2m | 2048 | User-derived: user id | User table | User save refreshes key | Sanitized base user snapshot without password hashes, reused by user show, user cards, and permission role lookup. |
| User public profile | `userservice` | `userservice.UserPublicProfile` | 2m | 1024 | User-derived: user id | Public user fields, statistics, user badges | User changes, badge changes, stat-affecting actions invalidate; activity flush updates cached last-active time when present; global badge definition changes clear public profile cache | Public display snapshot used to derive hover and full user cards; excludes email, password hashes, and account status fields. |
| User recent activity | `userservice` | Recent activity timestamp with last-flushed marker | 2m | 8192 | User-derived: user id | Authenticated request activity events | Refreshed on each activity event; per-user DB flush after 2m of continuous activity; expiration/capacity eviction and shutdown flush final activity to DB | Immediate online-status source; DB last-active is persistence rather than the hot path. |
| Badge definitions | `badgeservice` | Admin badge definitions | 10m | 4 | Near-finite: one key | System definitions plus `badges` table | Badge save/delete clears definitions and user cards | Uses `localcache`; user badge records are not cached directly. |
| Unread status | `unreadservice` | Notification/message unread booleans | 2m | 2048 | User-derived: user id | Notifications, chat config | Notification/chat writes delete user key | Small UI state cache. |

## Key Space Classes

- Finite: the number of keys is fixed by code. These caches mainly need clear
  invalidation and sensible stale budgets.
- Near-finite: the number of keys is technically data-driven but expected to stay
  small, such as categories or badge definitions.
- Capped input: key parts come from requests, but code bounds the cacheable
  range. Article list caching only stores normalized sorts for the first 50
  pages; deeper pages bypass the cache.
- User-derived or DB-bounded: key count can grow with users or roles. These
  caches must keep short TTLs or explicit invalidation, and set a capacity that
  matches expected working-set size.
- Input-derived: key count can grow from request-controlled values. These need
  validation, normalization, TTL, and an explicit capacity.

## Current Duplication And Nesting

- Business in-process caches use `localcache` so values stay typed and share one
  cache facade.
- Article list VO caching is intentionally a presentation snapshot. It embeds
  article fields, author names and avatars, poster names and avatars, and
  category names. It is acceptable only because the TTL is short.
- Category list and category map are served from one category snapshot so the
  same DB source is not cached twice.
- User display data is split into sanitized user info and public profile
  snapshots. User info can support account-side reads and role lookup; public
  profile keeps only display fields plus statistics and badges. User show is
  derived from user info, while hover/full cards are derived from public profile
  snapshots, with recent activity overlaid in memory for online status. Article
  list VO still keeps its own short-lived user display snapshot and is allowed
  to lag for up to 5 seconds.
- SEO XML depends on articles, categories, and site settings but is intentionally
  TTL-only. If admin changes need immediate SEO refresh, add a dedicated
  `ClearSEOXMLCache` hook.

## Deferred Work

- `hotdataserve` still owns several cache domains, but files are split by
  config, site statistics, article list snapshots, categories, and article type.
  Moving those domains into separate packages can be done later if ownership
  needs to be stricter.
- Article list caches intentionally rely on a 5 second stale budget for publish,
  edit, reply, and status changes. Adding event-driven invalidation is possible,
  but it would trade simple short-lived snapshots for broader write-path coupling.
- SEO XML is TTL-only with a 10 second stale budget. Immediate refresh on
  article, category, or site-setting changes can be added later with a dedicated
  `ClearSEOXMLCache` hook.
- `gzipCache` remains a `sync.Map` response cache because it only stores finite
  embedded static responses in production. It is separate from business data
  caching and does not need TTL cleanup.

## Rules For New Caches

1. Name the owner and value shape before adding a cache.
2. Record the stale budget and invalidation path in this document.
3. Classify the key space as finite, near-finite, capped input, user-derived,
   DB-bounded, or input-derived.
4. Prefer `localcache` for typed business data and short-lived UI state.
5. For input-derived caches, normalize keys, validate inputs, and add a cap or
   clear reason why global capacity is enough.
6. Avoid caching a VO that embeds other mutable domains unless the TTL is short
   or all dependencies have clear invalidation hooks.

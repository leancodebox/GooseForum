# React navigation keep-alive record

- The React site mirrors the Vue `KeepAlive(max=10)` boundary for `home.index`, `category.index`, and `search.index`.
- React's built-in `Activity` keeps cached list component trees and their local state while they are hidden; no third-party keep-alive package is required.
- Each browser history entry receives a stable key and stores its own window scroll position. Back/forward navigation restores that entry after the cached tree becomes visible.
- Normal link navigation still resets to the top or honors a URL hash. Refresh preserves the current position.
- Cached list pages are managed as an LRU collection capped at ten entries. Non-list pages are unmounted normally so detail, settings, and messaging effects do not remain active.
- Real-browser verification covered a loaded home waterfall, topic navigation, and browser Back: the cached list remained loaded and returned to the saved non-zero scroll position instead of the top.

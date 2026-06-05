# Awesome Go Submission Notes

This note tracks the current submission shape for adding GooseForum to
awesome-go.

## Suggested Category

`Software Packages` -> `Other Software`

GooseForum is deployable forum software, not a web framework or a library.

## Suggested Entry

```md
- [GooseForum](https://github.com/leancodebox/GooseForum) - Forum software built with Go, Gin, GORM, Vue, and SQLite/MySQL support.
```

## PR Body Links

```md
Forge link: https://github.com/leancodebox/GooseForum
pkg.go.dev: https://pkg.go.dev/github.com/leancodebox/GooseForum
goreportcard.com: https://goreportcard.com/report/github.com/leancodebox/GooseForum
Coverage: https://app.codecov.io/gh/leancodebox/GooseForum
```

## Current Readiness

- Repository is public and MIT licensed.
- The repository has more than 5 months of history.
- `go.mod` is present at the repository root.
- SemVer releases are available, including `v0.2.14`.
- pkg.go.dev is reachable for `github.com/leancodebox/GooseForum`.
- Go Report Card currently reports `A+`.
- CI now uploads `coverage.out` as an artifact and attempts a Codecov upload.
- Local full-package coverage has been raised from 11.8% to about 15.5% by
  adding low-risk tests for shared helper packages.

## Remaining Review Risk

- Current total Go test coverage is low, about 15.5% from a local full-package
  coverage run.
- Several public Go packages and symbols still have sparse or Chinese-only
  pkg.go.dev comments.
- Codecov needs to be enabled for the public repository after the workflow lands.

## Coverage Target

GooseForum is a deployable application rather than a reusable Go library, so
100% repository coverage is not the practical goal. The submission target is:

- Short term: raise full-package coverage above 20% while covering security,
  token, cache, URL, Markdown, resource, and configuration helpers.
- Medium term: reach roughly 25%-35% by adding focused tests around public HTTP
  flows and core services.
- Avoid brittle tests around CLI glue, external integrations, generated-style
  model repositories, and file watcher behavior unless they protect a real bug.

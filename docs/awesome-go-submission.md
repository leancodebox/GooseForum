# Awesome Go Submission Notes

This note records the practical plan for submitting GooseForum to
`avelino/awesome-go` / `awesome-go.com`.

## Current Conclusion

GooseForum already satisfies the main automatic checks for an awesome-go PR:

- Public repository: `https://github.com/leancodebox/GooseForum`
- Root `go.mod`: present.
- License: MIT.
- Project age: first local commit is `2023-04-26`, so it is older than 5 months.
- SemVer release: latest local tag is `v0.2.14`.
- pkg.go.dev link: `https://pkg.go.dev/github.com/leancodebox/GooseForum`
- Go Report Card link: `https://goreportcard.com/report/github.com/leancodebox/GooseForum`
- Coverage link: `https://app.codecov.io/gh/leancodebox/GooseForum`

The main risk is not entry formatting. The main risk is review judgment:
awesome-go is curated for high-quality Go packages/projects, and its guideline
mentions high test coverage when a project is testable. GooseForum is a
deployable forum application, so repository-wide coverage is naturally dragged
down by controllers, model repositories, CLI glue, and integration-heavy code.
Still, low coverage may receive a `needs-coverage` label.

## Recommendation

Do not delay the submission only to chase arbitrary coverage or add superficial
package comments.

Recommended approach:

1. Submit once Codecov shows the latest default-branch coverage report.
2. Keep the PR small: change only `README.md` in awesome-go.
3. Be explicit that GooseForum is deployable forum software, not a reusable Go
   library.
4. If reviewers ask for coverage, respond with targeted improvements around
   stable core logic instead of trying to cover generated-style repositories or
   external integrations.

## Suggested Category

`Software Packages` -> `Other Software`

Reason: GooseForum is end-user forum software. It is not a web framework, ORM,
middleware, library, or development tool.

## Alphabetical Position

In the current awesome-go `Other Software` list, GooseForum should be inserted
after:

```md
- [GoDocTooltip](https://github.com/diankong/GoDocTooltip) - Chrome extension for Go Doc sites, which shows function description as tooltip at function list.
```

and before:

```md
- [Gokapi](https://github.com/Forceu/gokapi) - Lightweight server to share files, which expire after a set amount of downloads or days. Similar to Firefox Send, but without public upload.
```

## Suggested Entry

Use a concise, non-promotional description that ends with a period:

```md
- [GooseForum](https://github.com/leancodebox/GooseForum) - Forum software with a Go backend, Vue frontend, and SQLite/MySQL support.
```

Alternative if reviewers prefer less frontend detail:

```md
- [GooseForum](https://github.com/leancodebox/GooseForum) - Forum software with SQLite and MySQL support.
```

## PR Body

Use the required links:

```md
Forge link: https://github.com/leancodebox/GooseForum
pkg.go.dev: https://pkg.go.dev/github.com/leancodebox/GooseForum
goreportcard.com: https://goreportcard.com/report/github.com/leancodebox/GooseForum
Coverage: https://app.codecov.io/gh/leancodebox/GooseForum
```

If a short note is useful, keep it factual:

```md
GooseForum is submitted under Software Packages / Other Software because it is
deployable forum software rather than a framework or library.
```

## Pre-Submission Checklist

- Confirm the public GitHub release page shows the latest SemVer release.
- Confirm pkg.go.dev opens for `github.com/leancodebox/GooseForum`.
- Confirm Go Report Card is `A-` or better.
- Confirm Codecov has a reachable report for the default branch.
- Confirm the awesome-go PR changes only `README.md`.
- Confirm the entry is alphabetically placed between `GoDocTooltip` and
  `Gokapi`.
- Confirm the entry ends with a period.

## Local Verification Commands

Run these before cutting a release or opening the PR:

```sh
gofmt -l $(rg --files -g '*.go')
go vet ./...
go test ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -n 1
go run github.com/client9/misspell/cmd/misspell@latest -error $(git ls-files)
```

Current local full-package coverage after recent focused tests:

```text
total: (statements) 16.7%
```

Current local lint status:

- `gofmt -l`: clean.
- `go vet ./...`: clean.
- `ineffassign ./...`: clean.
- `misspell` against tracked files: clean.
- `staticcheck ./...`: still reports legacy/generated-style model constants,
  plus a small number of style warnings that are not worth changing only for
  this submission.
- `gocyclo -over 20`: currently reports two higher-complexity functions:
  `SaveImgByGinContext` and `buildReplyPayloads`.

## Coverage Strategy

Coverage is worth improving only where tests protect real behavior.

Good candidates:

- Markdown rendering and search-content extraction.
- JWT, password, token, and auth helper edge cases.
- Cache behavior and closer lifecycle.
- Notification payload and unread-status behavior.
- Stable HTTP read flows with isolated test data.
- User/profile transformation logic.

Avoid chasing coverage in:

- Generated-style GORM repository methods.
- CLI command wiring with broad filesystem side effects.
- External OAuth, SMTP, captcha, Meilisearch, and storage integrations.
- File watcher and live-reload paths.

## If Reviewers Push Back

Likely labels/comments:

- `needs-coverage`: acknowledge that GooseForum is an application and continue
  adding targeted tests around stable core behavior.
- `needs-info`: verify that PR body links are present and reachable.
- Category concern: explain that existing entries in `Other Software` include
  deployable applications, and GooseForum is closest to that group.

Do not argue that coverage is irrelevant. The better position is:

```text
GooseForum is deployable software rather than a reusable package. I am improving
coverage around stable core behavior and avoiding brittle tests for external
integrations or generated-style persistence methods.
```

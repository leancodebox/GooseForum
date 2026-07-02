# Markdown Rendering Direction

GooseForum will keep Markdown as the canonical authoring format and use a
dual-implementation rendering strategy:

- Server rendering is authoritative for saved and public content.
- Client rendering is used for editor preview and progressive enhancement.
- The two implementations are kept aligned by a shared test specification, not
  by embedding a JavaScript runtime in the Go server.

This keeps the runtime simple while still leaving room for richer client-side
rendering such as diagrams, math, and code highlighting.

## Goals

- Store user-authored Markdown as the source of truth.
- Render final article and reply HTML on the server with `goldmark`.
- Render editor preview on the client with `markdown-it`.
- Define a Markdown compatibility test suite that both renderers must satisfy.
- Add advanced visual features as client-side enhancements when they are needed.
- Do not render raw HTML from user Markdown as a supported contract.

## Non-Goals

- Do not store rich-text HTML as the canonical content format.
- Do not add Vditor, WangEditor, or another full editor runtime by default.
- Do not embed Node, `goja`, or another JavaScript runtime in the Go server for
  normal Markdown rendering.
- Do not promise full compatibility with Discourse Markdown extensions.

## Rendering Model

```text
Editor raw Markdown
        |
        | client preview
        v
markdown-it preview HTML

Editor raw Markdown
        |
        | submit / save
        v
goldmark server HTML
        |
        | persisted rendered_html / page payload
        v
public article and reply display
```

The server result is the trusted result. Client preview should be close enough
for writing, but it is not the security or storage boundary.

## Markdown Compatibility Spec

The project should maintain a small corpus of Markdown fixtures under
`testdata/markdown-compat/`. Each fixture is a themed Markdown file with a
matching assertion file:

```text
testdata/markdown-compat/
  blocks.md
  blocks.json
  inline.md
  inline.json
```

Each fixture contains:

- A Markdown input.
- Expected semantic behavior.
- Allowed HTML differences when exact HTML is not important.

The suite should cover common forum content first:

- headings and generated heading IDs
- paragraphs and hard/soft line breaks
- emphasis, strong, strikethrough, inline code
- fenced code blocks
- ordered and unordered lists
- nested lists
- task lists
- blockquotes
- links and autolinks
- images
- tables
- raw HTML handling, currently treated as unsupported user markup
- escaped Markdown characters
- mixed Chinese and English punctuation

Exact HTML should only be asserted where the public contract depends on it. For
example, task list checkbox attributes and heading anchors matter; incidental
attribute order does not.

Current checks:

```bash
go test ./app/http/controllers/markdown2html
cd resource && pnpm exec vitest run test/markdown-compat.test.ts
```

## Server Responsibilities

Server rendering owns:

- final article and reply HTML
- security-sensitive normalization
- link attributes such as `target` and `rel`
- image attributes such as lazy loading and async decoding
- rendered version tracking and rebuild migrations
- SEO and no-JavaScript output

The current server renderer is `goldmark` in
`app/http/controllers/markdown2html`.

## Client Responsibilities

Client rendering owns:

- editor preview
- lightweight authoring helpers
- optional post-render enhancements that do not change stored Markdown

The current client preview renderer is centralized in
`resource/src/runtime/markdown.ts`. Pages should call this helper instead of
creating local `MarkdownIt` instances.

Potential client enhancements:

- Mermaid for diagrams
- KaTeX or MathJax for math
- Prism.js or another highlighter for code blocks

These should be loaded only on pages that need them, preferably by detecting
matching code fences or inline markers. They should not become part of the base
forum bundle until real usage justifies it.

Client enhancement libraries should decorate already-rendered content. They
should not change the canonical Markdown storage format.

## Feature Policy

Add Markdown features in this order:

1. Add or update compatibility fixtures.
2. Make `goldmark` pass the server expectations.
3. Make `markdown-it` preview match the semantic behavior.
4. Add client enhancement only if static HTML is not enough.
5. Document any accepted preview/final differences.

This avoids turning editor behavior into the content contract by accident.

## Current Decision

GooseForum should continue with dual implementation:

- `goldmark` for authoritative server rendering.
- `markdown-it` for client preview.
- fixture-based compatibility tests to keep them aligned.
- client-only optional renderers for diagrams, math, and code highlighting.

The `goja` experiment is useful as a reference, but it should not replace the
current approach unless the Markdown dialect becomes too complex to keep aligned
with tests.

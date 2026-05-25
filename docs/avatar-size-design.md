# Avatar Size Design

This document records the current avatar multi-size plan for GooseForum.

## Background

The current avatar upload flow crops the user-selected image in the browser and exports a single `400 x 400` image. The backend stores that uploaded file as-is and writes the stored file name to `users.avatar_url`.

That means every avatar usage currently loads the same `400 x 400` image, even when the UI only displays it at `24 x 24`, `32 x 32`, or `40 x 40`.

## Current Display Sizes

Current frontend avatar display sizes:

```text
Home topic list
- Mobile participants: 24 x 24
- Desktop participants: 32 x 32

Top navigation
- Current user avatar: 36 x 36

Article detail
- Title meta author: 20 x 20
- Article author: 44 x 44
- Reply author: 36 x 36, 40 x 40 on sm+
- Right rail participants: 32 x 32

User hover card
- Main avatar: 56 x 56

User profile
- Profile header avatar: 80 x 80, 96 x 96 on sm+
- Following/follower list: 40 x 40

User settings
- Side profile avatar: 48 x 48
- Upload preview avatar: 96 x 96
- Crop preview: 128 x 128

Admin
- Post management author avatar: 28 x 28
- User management list avatar: 36 x 36 / 40 x 40
- User edit dialog avatar: 44 x 44
```

## Decision

Generate three avatar files from the browser-side crop canvas:

```text
48 x 48
96 x 96
400 x 400
```

The backend should not resize, recompress, or transcode avatars. It should validate and store the files it receives.

This keeps GooseForum's deployment simple and avoids adding server-side image processing dependencies.

## Why Frontend Generation

The frontend already owns avatar crop interaction and already produces a high-quality square canvas. Generating smaller sizes from that same canvas is simple and avoids a backend image pipeline.

Benefits:

- No system dependencies such as libvips or ImageMagick.
- No new Go image processing dependency for the first phase.
- Browser canvas can already export WebP when supported.
- The backend remains a storage and validation layer.
- The generated files are deterministic because all sizes come from the same crop.

Tradeoffs:

- OAuth/external avatars are not automatically multi-sized in the first phase.
- The upload API needs to accept multiple avatar files.
- Older users without `avatar_key` must keep using the existing `avatar_url` fallback.

## File Naming

Use one size-less avatar key and derive file names from it.

Example for user `123`:

```text
avatar_key = avatars/u/123/avatar_1716610000_abcd

avatars/u/123/avatar_1716610000_abcd_48.webp
avatars/u/123/avatar_1716610000_abcd_96.webp
avatars/u/123/avatar_1716610000_abcd_400.webp
```

The exact suffix can use timestamp plus a short random token to avoid cache collisions.

## Database Design

Keep the existing field:

```text
users.avatar_url
```

Add a new field:

```text
users.avatar_key
```

Semantics:

```text
avatar_url  Compatibility field. Points to the default avatar, currently the 96px version.
avatar_key  Size-less avatar resource key. Used to derive 48/96/400 URLs.
```

Do not add one database column per size. Size policy belongs in code, not in the user table schema.

## API Shape

Keep the old field:

```json
{
  "avatarUrl": "/file/img/avatars/u/123/avatar_1716610000_abcd_96.webp"
}
```

Add multi-size URLs:

```json
{
  "avatarUrls": {
    "small": "/file/img/avatars/u/123/avatar_1716610000_abcd_48.webp",
    "medium": "/file/img/avatars/u/123/avatar_1716610000_abcd_96.webp",
    "large": "/file/img/avatars/u/123/avatar_1716610000_abcd_400.webp"
  }
}
```

Fallback rule:

```text
If avatar_key exists:
  derive avatarUrls from avatar_key
  avatarUrl = avatarUrls.medium

If avatar_key is empty but avatar_url exists:
  avatarUrl = current avatar_url
  avatarUrls.small = avatar_url
  avatarUrls.medium = avatar_url
  avatarUrls.large = avatar_url

If both are empty:
  use the default avatar for all fields
```

## Frontend Usage Rules

Use sizes like this:

```text
Displayed size <= 48px:
  avatarUrls.small

Displayed size > 48px and <= 96px:
  avatarUrls.medium

Large preview or future high-resolution display:
  avatarUrls.large
```

Current mapping:

```text
Home topic list participants: small
Top navigation current user: small
Article title meta author: small
Article author: small
Reply author: small
Article right rail participants: small
Admin list avatars: small

User hover card: medium
User profile header: medium
User settings avatar: medium

Crop preview / future avatar preview dialog: large
```

## Upload Flow

Current upload flow:

```text
User selects image
Browser opens crop modal
Cropper produces a 400 x 400 canvas
Frontend exports one avatar file
Backend stores the file
users.avatar_url is updated
```

Target upload flow:

```text
User selects image
Browser opens crop modal
Cropper produces a 400 x 400 canvas
Frontend exports:
  avatar_48.webp
  avatar_96.webp
  avatar_400.webp
Frontend uploads all three files in one request
Backend validates and stores all three files
Backend writes:
  users.avatar_key
  users.avatar_url = 96px avatar URL/key
API returns avatarUrl and avatarUrls
```

The avatar upload flow should not run `processImageFile` after `canvasToImageFile`. The canvas export is already an encode step; re-compressing introduces quality loss without clear benefit.

## Backend Responsibilities

The backend should:

- Require login.
- Validate that all required sizes are present.
- Validate file type and size.
- Store files under a shared avatar key.
- Update `users.avatar_key`.
- Keep `users.avatar_url` compatible by pointing it to the 96px avatar.
- Return `avatarUrl` and `avatarUrls`.

The backend should not:

- Resize avatars in the first phase.
- Recompress avatars.
- Depend on libvips, ImageMagick, or similar system packages.

## Migration Strategy

Phase 1:

- Add `users.avatar_key`.
- Update avatar upload API to accept 48/96/400 files.
- Keep `avatar_url` working.
- Add `avatarUrls` to user payloads.
- Update high-frequency avatar UI to use `avatarUrls.small`.

Phase 2:

- Move user profile and settings pages to `avatarUrls.medium`.
- Keep fallback for old avatars.
- Optionally add an old-avatar cleanup job.

Phase 3:

- Evaluate OAuth avatar handling.
- Consider backfilling generated sizes only if there is a strong need.

## Open Questions

- Whether old uploaded avatars should be kept forever or cleaned after successful replacement.
- Whether OAuth avatars should be copied and multi-sized, or left as external URLs.
- Whether `400` should be named `large` or `original` in public payloads. For now, use `large` because it is a cropped derivative, not the user's original upload.

import assert from 'node:assert/strict';
import test from 'node:test';
import { localLinkTargets } from './check-doc-links.mjs';

test('keeps real local links, but not fenced gallery syntax examples', () => {
  const source = '[Guide](guide.md#section)\n```markdown\n![Example](/path/one.jpg)\n```\n[Missing](missing.md)';
  assert.deepEqual(localLinkTargets(source), ['guide.md#section', 'missing.md']);
});

test('handles tilde fences and longer fences containing short fence examples', () => {
  const source = '~~~~markdown\n```\n[example](fake.md)\n```\n~~~~\n![Real image](image.png)';
  assert.deepEqual(localLinkTargets(source), ['image.png']);
});

test('ignores external URLs and anchors, and retains relative links', () => {
  assert.deepEqual(localLinkTargets('[Web](https://example.com) [Here](#section) [Local](../README.md)'), ['../README.md']);
});

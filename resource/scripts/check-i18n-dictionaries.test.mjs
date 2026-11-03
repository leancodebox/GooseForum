import assert from "node:assert/strict";
import test from "node:test";
import { inspectDictionary } from "./check-i18n-dictionaries.mjs";

test("detects duplicate quoted, unquoted and nested TypeScript dictionary keys", () => {
  const result = inspectDictionary(`export default { name: "first", "name": "second", nested: { a: "a", a: "b" } } as const;`);
  assert.deepEqual(result.errors, ["重复 key: name", "重复 key: a"]);
});

test("detects JSON duplicate keys before JSON.parse discards them", () => {
  assert.deepEqual(inspectDictionary('{"a":"first","a":"second"}', "en.json").errors, ["重复 key: a"]);
});

test("tracks explicit English fallback without treating overrides as duplicate keys", () => {
  const result = inspectDictionary(`import en from "./en-publish.js"; export default { ...en, name: "Translated" } as const;`);
  assert.deepEqual(result.errors, []);
  assert.equal(result.englishFile, "./en-publish.js");
  assert.equal(result.strings.get("name"), "Translated");
});

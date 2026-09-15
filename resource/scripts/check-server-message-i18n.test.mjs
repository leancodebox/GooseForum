import assert from "node:assert/strict";
import test from "node:test";
import {
  parseBackendMessageCodes,
  parseTranslationKeys,
  validateServerMessages,
} from "./check-server-message-i18n.mjs";

test("extracts Go MessageCode constants and TypeScript dictionary keys", () => {
  assert.deepEqual(
    parseBackendMessageCodes('const MessageEmail MessageCode = "auth.email.exists"'),
    ["auth.email.exists"],
  );
  assert.deepEqual(
    parseTranslationKeys('export default {\n  "auth.email.exists": "used",\n}'),
    ["auth.email.exists"],
  );
});

test("reports missing, stale, and duplicate translations", () => {
  assert.deepEqual(
    validateServerMessages(
      ["auth.email.exists", "auth.username.exists"],
      { zh: ["auth.email.exists", "auth.email.exists", "stale.key"] },
    ),
    [
      "zh 缺少: auth.username.exists",
      "zh 存在后端未声明的 key: stale.key",
      "zh 存在重复 key: auth.email.exists",
    ],
  );
});

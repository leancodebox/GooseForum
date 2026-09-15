import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";

const backendFile = fileURLToPath(
  new URL("../../app/http/controllers/component/message_code.go", import.meta.url),
);
const locales = ["zh", "en", "ja", "it"];

export function parseBackendMessageCodes(source) {
  return collectMatches(source, /\b\w+\s+MessageCode\s*=\s*"([^"]+)"/g);
}

export function parseTranslationKeys(source) {
  return collectMatches(source, /^\s*"([^"]+)"\s*:/gm);
}

export function validateServerMessages(backendCodes, translations) {
  const errors = [];
  const expected = new Set(backendCodes);
  const duplicateBackendCodes = duplicates(backendCodes);
  if (duplicateBackendCodes.length) {
    errors.push(`后端存在重复 MessageCode: ${duplicateBackendCodes.join(", ")}`);
  }

  for (const [locale, keys] of Object.entries(translations)) {
    const actual = new Set(keys);
    const missing = [...expected].filter((key) => !actual.has(key)).sort();
    const stale = [...actual].filter((key) => !expected.has(key)).sort();
    const duplicateKeys = duplicates(keys);
    if (missing.length) errors.push(`${locale} 缺少: ${missing.join(", ")}`);
    if (stale.length) errors.push(`${locale} 存在后端未声明的 key: ${stale.join(", ")}`);
    if (duplicateKeys.length) errors.push(`${locale} 存在重复 key: ${duplicateKeys.join(", ")}`);
  }
  return errors;
}

function collectMatches(source, pattern) {
  return [...source.matchAll(pattern)].map((match) => match[1]);
}

function duplicates(values) {
  const seen = new Set();
  const result = new Set();
  for (const value of values) {
    if (seen.has(value)) result.add(value);
    seen.add(value);
  }
  return [...result].sort();
}

async function main() {
  const backendCodes = parseBackendMessageCodes(await readFile(backendFile, "utf8"));
  const sources = Object.fromEntries(
    await Promise.all(
      locales.map(async (locale) => {
        const file = fileURLToPath(
          new URL(
            `../packages/client/src/i18n/messages/${locale}-server-messages.ts`,
            import.meta.url,
          ),
        );
        return [locale, await readFile(file, "utf8")];
      }),
    ),
  );
  const englishKeys = parseTranslationKeys(sources.en);
  const translations = Object.fromEntries(
    locales.map((locale) => {
      const keys = parseTranslationKeys(sources[locale]);
      return [
        locale,
        locale !== "en" && sources[locale].includes("...en,")
          ? [...new Set([...englishKeys, ...keys])]
          : keys,
      ];
    }),
  );
  const errors = validateServerMessages(backendCodes, translations);
  if (errors.length) {
    console.error("后端 MessageCode 与前端翻译不一致：");
    for (const error of errors) console.error(`- ${error}`);
    process.exitCode = 1;
    return;
  }
  console.log(
    `Server message i18n check passed: ${backendCodes.length} keys × ${locales.length} locales.`,
  );
}

if (process.argv[1] === fileURLToPath(import.meta.url)) await main();

/// <reference types="node" />

import { readdirSync, readFileSync } from "node:fs";
import { join, resolve } from "node:path";
import { beforeAll, describe, expect, it } from "vitest";
import {
  adminAuditMessageCodes,
  formatAuditMessage,
} from "./audit-i18n";
import {
  getAdminDictionary,
  prepareAdminTranslations,
} from "./translation-loader";

beforeAll(async () => {
  await Promise.all(
    (["zh", "en", "ja", "it"] as const).map((locale) =>
      prepareAdminTranslations(locale, ["audit", "audit-extra"]),
    ),
  );
});

describe("admin audit i18n", () => {
  it("localizes structured messages, statuses, and changed fields", () => {
    expect(
      formatAuditMessage(
        "zh",
        "admin.opt.user.updated",
        {
          userId: 42,
          changes: ["status", "activation", "role"],
        },
        "",
      ),
    ).toBe("更新用户 42：账号状态, 验证状态, 角色");

    expect(
      formatAuditMessage(
        "zh",
        "admin.opt.topic.statusChanged",
        {
          title: "公告",
          status: "blocked",
        },
        "",
      ),
    ).toBe("主题「公告」状态调整为封禁");
  });

  it("falls back to the original log text for unknown message codes", () => {
    expect(
      formatAuditMessage("en", "plugin.unknown", {}, "original text"),
    ).toBe("original text");
  });

  it("keeps every audit dictionary structurally complete across locales", () => {
    const locales = ["zh", "en", "ja", "it"] as const;
    const baseKeys = sortedKeys(getAdminDictionary("audit", "en"));
    const extra = getAdminDictionary<Record<string, Record<string, unknown>>>(
      "audit-extra",
      "en",
    );
    for (const locale of locales) {
      expect(sortedKeys(getAdminDictionary("audit", locale))).toEqual(baseKeys);
      const dictionary = getAdminDictionary<
        Record<string, Record<string, unknown>>
      >("audit-extra", locale);
      expect(sortedKeys(dictionary)).toEqual(sortedKeys(extra));
      for (const group of sortedKeys(extra))
        expect(sortedKeys(dictionary[group])).toEqual(sortedKeys(extra[group]));
      for (const key of sortedKeys(extra.messages))
        expect(templateParams(dictionary.messages[key])).toEqual(
          templateParams(extra.messages[key]),
        );
    }
  });

  it("localizes content review types and actions", () => {
    expect(
      formatAuditMessage(
        "zh",
        "admin.opt.content.reviewed",
        {
          type: "post",
          subjectId: 18,
          action: "reject",
          version: 3,
          reason: "spam",
        },
        "",
      ),
    ).toBe("审核回复 #18，操作拒绝，版本 3：spam");
    expect(
      formatAuditMessage(
        "ja",
        "admin.opt.content.reviewed",
        {
          type: "topic",
          subjectId: 9,
          action: "approve",
          version: 2,
          reason: "確認済み",
        },
        "",
      ),
    ).toBe("トピック #9 を審査：承認、バージョン 2：確認済み");
  });

  it.each([
    "admin.opt.content.reviewed",
    "admin.opt.user.updated",
    "admin.opt.topic.statusChanged",
    "admin.opt.topic.pinWeightChanged",
    "admin.opt.topic.categoriesChanged",
    "admin.opt.topic.deleted",
    "moderator.opt.topic.statusChanged",
    "admin.opt.category.moderatorAdded",
    "admin.opt.category.moderatorRemoved",
  ])("has a localized template for %s", (code) => {
    expect(formatAuditMessage("zh", code, {}, "UNMAPPED")).not.toBe("UNMAPPED");
  });

  it("covers every operation-log message code emitted by the Go backend", () => {
    const backendRoot = resolve(process.cwd(), "../../../app");
    const sources = collectGoFiles(backendRoot)
      .map((file) => readFileSync(file, "utf8"))
      .join("\n");
    const calls = sources.match(/optlogger\.UserOptCode\(/g) || [];
    const codes = [
      ...sources.matchAll(
        /optlogger\.UserOptCode\([\s\S]{0,320}?"((?:admin|moderator)\.opt\.[^"]+)"\s*,\s*optlogger\.MessageParams/g,
      ),
    ].map((match) => match[1]);

    expect(codes).toHaveLength(calls.length);
    expect(codes.length).toBeGreaterThan(0);
    for (const code of codes) expect(adminAuditMessageCodes).toHaveProperty(code);
    expect(
      Object.keys(adminAuditMessageCodes).filter((code) => !codes.includes(code)),
    ).toEqual(["moderator.opt.topic.statusChanged"]);
  });
});

function collectGoFiles(directory: string): string[] {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    return entry.isDirectory()
      ? collectGoFiles(path)
      : entry.isFile() && entry.name.endsWith(".go")
        ? [path]
        : [];
  });
}

function sortedKeys(value: object) {
  return Object.keys(value).sort();
}

function templateParams(value: unknown) {
  return typeof value === "string"
    ? [...value.matchAll(/\{(\w+)\}/g)].map((match) => match[1]).sort()
    : [];
}

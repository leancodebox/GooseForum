import { describe, expect, it } from "vitest";
import { formatAuditMessage } from "./audit-i18n";

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
});

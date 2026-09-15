import { expect, it } from "vitest";
import { withEditorId, withoutEditorId } from "./editor-identity";

it("keeps identity through cloning and reordering without sending it to APIs", () => {
  const first = withEditorId({ name: "same" });
  const second = withEditorId({ name: "same" });
  expect(first.editorId).not.toBe(second.editorId);
  const moved = structuredClone([second, first]);
  expect(withEditorId(moved[1]).editorId).toBe(first.editorId);
  expect(moved.map(withoutEditorId)).toEqual([
    { name: "same" },
    { name: "same" },
  ]);
});

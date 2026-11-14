import { describe, expect, it } from "vitest";
import { appendUniqueById, prependUniqueById } from "../src/site/collection-utils";

describe("collection utils", () => {
  it("appends only unseen ids and preserves order", () => {
    expect(appendUniqueById([{ id: 1 }, { id: 2 }], [{ id: 2 }, { id: 3 }, { id: 3 }])).toEqual([
      { id: 1 },
      { id: 2 },
      { id: 3 },
    ]);
  });

  it("prepends only unseen ids and preserves incoming order", () => {
    expect(prependUniqueById([{ id: 2 }, { id: 3 }], [{ id: 1 }, { id: 2 }, { id: 1 }])).toEqual([
      { id: 1 },
      { id: 2 },
      { id: 3 },
    ]);
  });
});

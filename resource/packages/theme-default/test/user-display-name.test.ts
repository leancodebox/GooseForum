import { expect, it } from "vitest";
import { userDisplayName } from "../src/site/users/display-name";

it("uses the nickname for display and falls back for missing or blank nicknames", () => {
  expect(userDisplayName({ username: "account", nickname: " Display name " })).toBe("Display name");
  for (const nickname of [undefined, "", " \t "]) {
    expect(userDisplayName({ username: "account", nickname })).toBe("account");
  }
});

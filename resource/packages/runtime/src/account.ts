import type { GooseSiteApi } from "@gooseforum/client";

export async function logout(api: GooseSiteApi, userId: number, redirect: () => void) {
  if (!window.dispatchEvent(new Event("goose:before-logout", { cancelable: true }))) return;
  await api.auth.logout();
  const { clearAccountDrafts, draftOwner } = await import("./local-draft-store");
  await clearAccountDrafts(draftOwner(userId));
  redirect();
}

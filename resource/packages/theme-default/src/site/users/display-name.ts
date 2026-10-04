export function userDisplayName(user: { username: string; nickname?: string }) {
  return user.nickname?.trim() || user.username;
}

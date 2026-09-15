// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-moderation-settings";
export default {
  ...en,
  http: "HTTP 通知",
  sensitive: "センシティブワード",
} as const;

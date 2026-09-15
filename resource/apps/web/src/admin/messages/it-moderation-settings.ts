// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-moderation-settings";
export default {
  ...en,
  http: "Notifiche HTTP",
  sensitive: "Parole sensibili",
} as const;

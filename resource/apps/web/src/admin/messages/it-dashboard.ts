// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-dashboard";
export default { ...en, title: "Dashboard" } as const;

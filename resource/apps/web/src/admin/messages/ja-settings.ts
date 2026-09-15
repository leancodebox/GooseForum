// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default { ...en, site: "サイト情報", chrome: "サイト表示" } as const;

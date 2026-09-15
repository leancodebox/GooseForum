// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-assets";
export default { ...en, badges: "Badge", files: "File" } as const;

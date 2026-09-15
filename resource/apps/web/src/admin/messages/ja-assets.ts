// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-assets";
export default { ...en, badges: "バッジ", files: "ファイル" } as const;

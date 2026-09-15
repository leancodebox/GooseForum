// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-system-settings";
export default { ...en, mail: "Email", security: "Sicurezza" } as const;

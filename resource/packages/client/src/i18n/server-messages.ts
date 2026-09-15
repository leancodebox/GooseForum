import zh from "./messages/zh-server-messages.js";
import en from "./messages/en-server-messages.js";
import ja from "./messages/ja-server-messages.js";
import it from "./messages/it-server-messages.js";
import type { Locale } from "./auth.js";

export const serverMessageResources = { zh, en, ja, it } as const;

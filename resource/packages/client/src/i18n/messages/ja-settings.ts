// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings.js";
export default {
  ...en,
  tabs: {
    sessions: "ログイン端末",
    profile: "プロフィール",
    account: "アカウント",
    privacy: "プライバシー",
    binding: "連携",
    applications: "認可済みアプリ",
  },
  save: "保存",
  cancel: "キャンセル",
  edit: "編集",
} as const;

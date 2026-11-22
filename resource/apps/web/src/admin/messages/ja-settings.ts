// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-settings";
export default {
  ...en,
  assignAvatar: "新規ユーザーに既定のアバターを割り当てる",
  assignAvatarHint: "アバターがない新規ユーザーが対象です。画像が未設定なら内蔵画像をランダムに割り当て、無効時は表示名の先頭文字を表示します。",
  defaultAvatar: "既定のアバター画像",
  removeAvatar: "画像を削除",
  site: "サイト情報",
  chrome: "サイト表示",
  showInHeader: "ヘッダーに{item}を表示",
  showInSidebar: "サイドバーに{item}を表示",
  more: "もっと見る",
  members: "メンバー",
  accessGroups: "アクセスグループ",
  themePreview: "テーマプレビュー",
} as const;

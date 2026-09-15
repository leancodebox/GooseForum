// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-settings";
export default {
  ...en,
  posting: "投稿設定",
  announcement: "お知らせ",
  announcementContentHint: "Markdown で作成し、公開時と同じ表示をプレビューできます。",
  announcementEditorPreview: "プレビュー",
  announcementEditorEmptyPreview: "プレビューする内容がありません。",
  announcementEditorUploadImage: "画像をアップロード",
  announcementEditorUploadFailed: "画像のアップロードに失敗しました",
  announcementEditorPlaceholder: "Markdown でお知らせを入力",
} as const;

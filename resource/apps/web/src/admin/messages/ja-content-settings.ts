// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-content-settings";
export default {
  ...en,
  externalLinks: "投稿の外部リンク",
  externalLinksEnabled: "外部サイトへの移動前に警告する",
  externalLinksHint: "トピック本文と返信のみに適用され、お知らせには適用されません。無効時と許可リスト内のドメインは直接移動します。",
  externalLinksWhitelist: "許可するドメイン",
  externalLinksWhitelistHint: "1行に1つのホスト名（example.com）。完全一致のみ。サブドメインは個別に指定し、プロトコル・パス・ポートは含めないでください。",
  posting: "投稿設定",
  announcement: "お知らせ",
  announcementContentHint: "Markdown で作成し、公開時と同じ表示をプレビューできます。",
  announcementEditorPreview: "プレビュー",
  announcementEditorEmptyPreview: "プレビューする内容がありません。",
  announcementEditorUploadImage: "画像をアップロード",
  announcementEditorUploadFailed: "画像のアップロードに失敗しました",
  announcementEditorPlaceholder: "Markdown でお知らせを入力",
} as const;

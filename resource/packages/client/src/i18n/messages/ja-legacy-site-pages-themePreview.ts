// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  presetsTitle: "プリセット",
  presetApplied:
    "{name} プリセットを適用しました。草稿を保存すると保持されます",
  saveDraft: "草稿を保存",
  publishSite: "全体に公開",
  workflowTitle: "使い方:",
  workflowDescription:
    "現在編集しているのはテーマ草稿です。草稿の保存だけではサイト全体には反映されません。公開後、左側の有効化スイッチがオンのときだけカスタムテーマが適用されます。",
  draftSaved: "テーマ草稿を保存しました。サイト全体にはまだ反映されません。",
  published: "テーマをサイト全体に公開しました。",
  saveFailed: "テーマ草稿の保存に失敗しました",
  publishFailed: "テーマの公開に失敗しました",
  restoredToSaved: "保存済みの設定に戻しました",
  themeRestoredDefault: "{name} を組み込みの既定に戻しました",
  allRestoredDefault:
    "組み込みの既定テーマに戻しました。プレ公開を保存すると全体に公開できます",
  cssCopied: "CSS をコピーしました",
  cssCopyFailed: "CSS のコピーに失敗しました",
  pageTitle: "テーマプレビュー設定",
  resetDefault: "既定",
  restoreToSavedTitle: "保存済みの設定に戻す",
  enableLabel: "有効化",
  tabLatest: "最新",
  tabHot: "人気",
  tabFeatured: "注目",
  samplePublish: "トピックを投稿",
  sampleTopic1:
    "テーマシステム再構築の議論：カラー、角丸、コンポーネントの状態",
  sampleTopic2: "ダークモードでの Markdown 本文の可読性",
  sampleTopic3: "新規ユーザーの導線と通知のビジュアル確認",
  sampleTopicExcerpt:
    "現在のテーマで、弱いテキスト・区切り線・タグ・ホバー背景の階層を確認します。",
  sampleTopicTitle: "あるトピック投稿のタイトル",
  sampleTopicBody:
    "本文・リンク・引用・コードブロックをここでシミュレートします。テーマ変数はライト／ダークの両モードで内容を読みやすく保つ必要があります。特に本文・境界線・弱いテキストです。",
  sampleTopicQuote:
    "カラーの階層は静かであるべきですが、曖昧にしてはいけません。",
  sampleMessageIncoming: "この背景は快適ですか？",
  sampleMessageOutgoing: "コントラストは安定させる必要があります。",
  sampleTextarea: "テーマ変数は入力・フォーカス・本文をカバーします。",
  presets: {
    goose: {
      label: "Goose",
      description: "現在の標準、すっきり安定",
    },
    clean: {
      label: "Clean",
      description: "プロダクト向けの低彩度ブルーとティール",
    },
    warm: {
      label: "Warm",
      description: "コミュニティ向けの柔らかい暖色",
    },
    deep: {
      label: "Deep",
      description: "黒を基調にした深いコントラスト",
    },
    cupcake: {
      label: "Cupcake",
      description: "やわらかく軽いピンク系",
    },
    retro: {
      label: "Retro",
      description: "気軽なコミュニティ向けの温かいレトロ調",
    },
    synthwave: {
      label: "Synthwave",
      description: "存在感のあるネオン紫とシアン",
    },
    aqua: {
      label: "Aqua",
      description: "明るく透明感のあるシアン系",
    },
    forest: {
      label: "Forest",
      description: "知識コミュニティ向けの深いグリーン",
    },
  },
} as const;

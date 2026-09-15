// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "{site} でログイン",
  subtitle: "アクセス許可のリクエスト",
  verifying: "リクエストを確認しています…",
  expired:
    "リクエストの有効期限が切れました。アプリに戻り、再度ログインしてください。",
  loadFailed: "リクエストを読み込めませんでした。",
  decisionFailed: "リクエストを処理できませんでした。",
  back: "{site} に戻る",
  accessAccount: "あなたの {site} アカウントへのアクセスを求めています",
  permissions: "許可すると、このアプリは次の操作ができます：",
  clientId: "クライアント ID",
  trust:
    "信頼できるアプリか確認してください。アクセス権はアカウント設定からいつでも取り消せます。",
  deny: "拒否",
  approve: "許可して続行",
  loading: "処理中…",
  scopes: {
    openid: "本人確認を行う",
    profile: "名前、ユーザー名、アバターを読み取る",
    email: "メールアドレスと確認状態を読み取る",
    offline_access: "利用していない間もログイン状態を維持する",
  },
} as const;

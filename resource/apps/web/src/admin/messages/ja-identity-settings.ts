// Keep the original English fallback live instead of duplicating its strings.
import en from "./en-identity-settings";
export default {
  ...en,
  oauth: "OAuth ログイン",
  oidc: "OIDC プロバイダー",
  retry: "再試行",
  loadFailedHint: "OAuth 設定を読み込めません。ネットワークまたはサーバーの状態を確認して再試行してください。",
  providerCheckFailed: "プロバイダー設定を確認できません。認証情報、Discovery URL、ネットワーク接続を確認してください。",
  callbackHint: "OAuth プロバイダーに完全なコールバック URL を登録してください。各部分にカーソルを合わせると生成元を確認できます。",
  callbackSiteHint: "サイト情報で設定されたサイト URL です。",
  callbackRouteHint: "GooseForum 固定の OAuth コールバックパスです。",
  callbackProviderHint: "現在のプロバイダーキーから生成されます。",
  callbackProviderPlaceholder: "provider-key/callback",
  copyCallback: "完全なコールバック URL をコピー",
  callbackCopied: "コールバック URL をコピーしました",
  copyFailed: "コピーに失敗しました",
  removeProvider: "プロバイダーを削除",
  removeProviderTitle: "この OAuth プロバイダーを削除しますか？",
  removeProviderHint: "設定を保存すると、このプロバイダーは削除されます：",
  addScope: "Scope を追加",
  removeScope: "Scope を削除",
} as const;

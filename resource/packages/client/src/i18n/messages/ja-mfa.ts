export default {
  title: '二要素認証', enabled: '有効', disabled: '無効',
  cancel: '登録をキャンセル',
  unavailable: '二要素認証を利用できません。管理者にお問い合わせください。',
  code: '認証アプリのコードまたはリカバリーコード', verify: '確認', enable: '認証アプリを登録', confirm: '二要素認証を有効にする',
  disable: '二要素認証を無効にする', regenerate: 'リカバリーコードを再生成', password: 'アカウントのパスワード',
  failed: '認証に失敗したか期限切れです。パスワードとコードをご確認ください。',
  statusFailed: '二要素認証設定の読み込みに失敗しました。', retry: '再試行',
  secret: '登録キー', qr: '認証アプリ登録用QRコード', recoveryTitle: 'リカバリーコード',
  recoveryNotice: '各コードは一度だけ使用できます。今すぐ保存してください。すべての端末からログアウトしました。',
  download: 'リカバリーコードをダウンロード', back: 'ログインに戻る', remaining: '残り {count} 個のリカバリーコード',
  logoutNotice: '変更するとすべての端末からログアウトします。',
  oauthPassword: '外部サービスでログインする場合は、プロフィールでメールを登録・確認し、パスワードのリセットで設定してください。',
} as const

import type { Locale } from './auth.js'

export interface OIDCConsentMessages {
  title: string
  subtitle: string
  verifying: string
  expired: string
  loadFailed: string
  decisionFailed: string
  back: string
  accessAccount: string
  permissions: string
  clientId: string
  trust: string
  deny: string
  approve: string
  loading: string
  scopes: Record<'openid' | 'profile' | 'email' | 'offline_access', string>
}

export const oidcConsentResources: Record<Locale, OIDCConsentMessages> = {
  zh: {
    title: '使用 {site} 登录', subtitle: '授权请求', verifying: '正在验证授权请求…', expired: '授权请求已失效，请返回应用重新发起登录。',
    loadFailed: '无法加载授权请求。', decisionFailed: '授权请求处理失败。', back: '返回 {site}', accessAccount: '希望访问你的 {site} 账号',
    permissions: '允许后，此应用可以：', clientId: '客户端 ID', trust: '请确认你信任此应用。你可以随时在账号设置中撤销访问权限。',
    deny: '拒绝', approve: '允许并继续', loading: '处理中…',
    scopes: { openid: '确认你的身份', profile: '读取名称、用户名和头像', email: '读取邮箱和验证状态', offline_access: '在你离开后继续保持登录' },
  },
  en: {
    title: 'Sign in with {site}', subtitle: 'Authorization request', verifying: 'Verifying authorization request…', expired: 'This request has expired. Return to the app and sign in again.',
    loadFailed: 'Unable to load the authorization request.', decisionFailed: 'Unable to process the authorization request.', back: 'Return to {site}', accessAccount: 'Wants to access your {site} account',
    permissions: 'This app will be allowed to:', clientId: 'Client ID', trust: 'Make sure you trust this app. You can revoke access in your account settings at any time.',
    deny: 'Deny', approve: 'Allow and continue', loading: 'Working…',
    scopes: { openid: 'Confirm your identity', profile: 'Read your name, username, and avatar', email: 'Read your email and verification status', offline_access: 'Keep you signed in while you are away' },
  },
  ja: {
    title: '{site} でログイン', subtitle: 'アクセス許可のリクエスト', verifying: 'リクエストを確認しています…', expired: 'リクエストの有効期限が切れました。アプリに戻り、再度ログインしてください。',
    loadFailed: 'リクエストを読み込めませんでした。', decisionFailed: 'リクエストを処理できませんでした。', back: '{site} に戻る', accessAccount: 'あなたの {site} アカウントへのアクセスを求めています',
    permissions: '許可すると、このアプリは次の操作ができます：', clientId: 'クライアント ID', trust: '信頼できるアプリか確認してください。アクセス権はアカウント設定からいつでも取り消せます。',
    deny: '拒否', approve: '許可して続行', loading: '処理中…',
    scopes: { openid: '本人確認を行う', profile: '名前、ユーザー名、アバターを読み取る', email: 'メールアドレスと確認状態を読み取る', offline_access: '利用していない間もログイン状態を維持する' },
  },
  it: {
    title: 'Accedi con {site}', subtitle: 'Richiesta di autorizzazione', verifying: 'Verifica della richiesta…', expired: 'La richiesta è scaduta. Torna all’app e accedi di nuovo.',
    loadFailed: 'Impossibile caricare la richiesta di autorizzazione.', decisionFailed: 'Impossibile elaborare la richiesta di autorizzazione.', back: 'Torna a {site}', accessAccount: 'Vuole accedere al tuo account {site}',
    permissions: 'Questa app potrà:', clientId: 'ID client', trust: 'Assicurati di fidarti di questa app. Puoi revocare l’accesso nelle impostazioni dell’account in qualsiasi momento.',
    deny: 'Rifiuta', approve: 'Consenti e continua', loading: 'Elaborazione…',
    scopes: { openid: 'Confermare la tua identità', profile: 'Leggere nome, nome utente e avatar', email: 'Leggere email e stato di verifica', offline_access: 'Mantenere l’accesso quando non sei presente' },
  },
}

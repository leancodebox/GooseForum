export const supportedLocales = ['zh', 'en', 'ja', 'it'] as const

export type Locale = (typeof supportedLocales)[number]

export interface AuthMessages {
  locale: { label: string, short: string }
  login: string
  register: string
  loginTitle: string
  registerTitle: string
  forgotTitle: string
  loginSubtitle: string
  registerSubtitle: string
  forgotSubtitle: string
  usernameOrEmail: string
  username: string
  email: string
  registeredEmail: string
  password: string
  newPassword: string
  confirmPassword: string
  captcha: string
  captchaAlt: string
  refreshCaptcha: string
  forgotPassword: string
  agreeTerms: string
  createAccount: string
  sendResetEmail: string
  backToLogin: string
  continueWith: string
  resetPasswordTitle: string
  resetPasswordSubtitle: string
  resetMissingToken: string
  passwordMinLength: string
  saveNewPassword: string
  passwordAdviceTitle: string
  passwordAdviceDescription: string
  passwordAdvice: { length: string, unique: string, loginAfterReset: string }
  validation: {
    loginRequired: string
    registerRequired: string
    forgotRequired: string
    passwordMismatch: string
    termsRequired: string
    captchaLoadFailed: string
    loginFailed: string
    registerFailed: string
    registerSuccess: string
    resetEmailFailed: string
  }
  server: { passwordResetMailQueued: string, passwordResetSuccess: string, passwordResetFailed: string }
}

const zh: AuthMessages = {
  locale: { label: '简体中文', short: '中' }, login: '登录', register: '注册',
  loginTitle: '登录账号', registerTitle: '创建新账号', forgotTitle: '重置密码',
  loginSubtitle: '欢迎回来，继续你的讨论和创作。', registerSubtitle: '加入 GooseForum，开启你的讨论空间。',
  forgotSubtitle: '输入邮箱，我们会发送一封重置密码邮件。', usernameOrEmail: '用户名或邮箱', username: '用户名',
  email: '邮箱', registeredEmail: '注册邮箱', password: '密码', newPassword: '新密码', confirmPassword: '确认密码', captcha: '验证码',
  captchaAlt: '验证码', refreshCaptcha: '刷新验证码', forgotPassword: '忘记密码？',
  agreeTerms: '我已阅读并同意服务条款和隐私政策', createAccount: '创建账号', sendResetEmail: '发送重置邮件',
  backToLogin: '返回登录', continueWith: '或继续使用',
  resetPasswordTitle: '重置密码', resetPasswordSubtitle: '设置一个新的登录密码。提交后即可返回登录页使用新密码登录。',
  resetMissingToken: '重置链接缺少 token，请重新从邮件打开。', passwordMinLength: '密码长度至少 6 位', saveNewPassword: '保存新密码',
  passwordAdviceTitle: '密码安全建议', passwordAdviceDescription: '设置一个只用于 GooseForum 的新密码。重置链接只在有效期内可用，过期后请重新发送邮件。',
  passwordAdvice: { length: '至少 6 位字符', unique: '避免与其他网站共用密码', loginAfterReset: '重置成功后返回登录页使用新密码' },
  validation: {
    loginRequired: '请填写账号、密码和验证码', registerRequired: '请完整填写注册信息', forgotRequired: '请填写邮箱和验证码',
    passwordMismatch: '两次输入的密码不一致', termsRequired: '请先同意服务条款和隐私政策', captchaLoadFailed: '验证码加载失败',
    loginFailed: '登录失败', registerFailed: '注册失败', registerSuccess: '注册成功', resetEmailFailed: '重置邮件发送失败',
  },
  server: { passwordResetMailQueued: '操作成功：如果该邮箱已注册，您将收到密码重置邮件', passwordResetSuccess: '密码重置成功', passwordResetFailed: '重置密码失败' },
}

const en: AuthMessages = {
  locale: { label: 'English', short: 'EN' }, login: 'Log in', register: 'Sign up',
  loginTitle: 'Log in to your account', registerTitle: 'Create an account', forgotTitle: 'Reset password',
  loginSubtitle: 'Welcome back. Continue your discussions and writing.', registerSubtitle: 'Join GooseForum and start your discussion space.',
  forgotSubtitle: 'Enter your email and we will send a password reset message.', usernameOrEmail: 'Username or email', username: 'Username',
  email: 'Email', registeredEmail: 'Registered email', password: 'Password', newPassword: 'New password', confirmPassword: 'Confirm password', captcha: 'Captcha',
  captchaAlt: 'Captcha', refreshCaptcha: 'Refresh captcha', forgotPassword: 'Forgot password?',
  agreeTerms: 'I have read and agree to the terms and privacy policy', createAccount: 'Create account', sendResetEmail: 'Send reset email',
  backToLogin: 'Back to login', continueWith: 'Or continue with',
  resetPasswordTitle: 'Reset password', resetPasswordSubtitle: 'Set a new login password. After submitting, return to log in with it.',
  resetMissingToken: 'The reset link is missing a token. Please open it again from your email.', passwordMinLength: 'Password must be at least 6 characters', saveNewPassword: 'Save new password',
  passwordAdviceTitle: 'Password safety tips', passwordAdviceDescription: 'Set a new password used only for GooseForum. Reset links are valid for a limited time; request a new email after expiration.',
  passwordAdvice: { length: 'At least 6 characters', unique: 'Avoid reusing passwords from other sites', loginAfterReset: 'After resetting, return to log in with the new password' },
  validation: {
    loginRequired: 'Please enter account, password, and captcha', registerRequired: 'Please complete the registration form', forgotRequired: 'Please enter email and captcha',
    passwordMismatch: 'The two passwords do not match', termsRequired: 'Please agree to the terms and privacy policy first', captchaLoadFailed: 'Failed to load captcha',
    loginFailed: 'Login failed', registerFailed: 'Registration failed', registerSuccess: 'Registration successful', resetEmailFailed: 'Failed to send reset email',
  },
  server: { passwordResetMailQueued: 'If this email is registered, you will receive a password reset email', passwordResetSuccess: 'Password reset successful', passwordResetFailed: 'Failed to reset password' },
}

const ja: AuthMessages = {
  locale: { label: '日本語', short: '日' }, login: 'ログイン', register: '登録',
  loginTitle: 'アカウントにログイン', registerTitle: '新規アカウント作成', forgotTitle: 'パスワード再設定',
  loginSubtitle: 'おかえりなさい。議論と投稿を続けましょう。', registerSubtitle: 'GooseForum に参加して、議論の場を始めましょう。',
  forgotSubtitle: 'メールアドレスを入力すると、再設定メールを送信します。', usernameOrEmail: 'ユーザー名またはメール', username: 'ユーザー名',
  email: 'メール', registeredEmail: '登録メール', password: 'パスワード', newPassword: '新しいパスワード', confirmPassword: 'パスワード確認', captcha: '認証コード',
  captchaAlt: '認証コード', refreshCaptcha: '認証コードを更新', forgotPassword: 'パスワードを忘れましたか？',
  agreeTerms: '利用規約とプライバシーポリシーに同意します', createAccount: 'アカウント作成', sendResetEmail: '再設定メールを送信',
  backToLogin: 'ログインへ戻る', continueWith: 'または次で続行',
  resetPasswordTitle: 'パスワード再設定', resetPasswordSubtitle: '新しいログインパスワードを設定します。送信後、新しいパスワードでログインできます。',
  resetMissingToken: '再設定リンクにトークンがありません。メールからもう一度開いてください。', passwordMinLength: 'パスワードは6文字以上にしてください', saveNewPassword: '新しいパスワードを保存',
  passwordAdviceTitle: 'パスワード安全のヒント', passwordAdviceDescription: 'GooseForum 専用の新しいパスワードを設定してください。再設定リンクは有効期限内のみ利用できます。',
  passwordAdvice: { length: '6文字以上', unique: '他のサイトと同じパスワードを使わない', loginAfterReset: '再設定後は新しいパスワードでログインしてください' },
  validation: {
    loginRequired: 'アカウント、パスワード、認証コードを入力してください', registerRequired: '登録情報をすべて入力してください', forgotRequired: 'メールと認証コードを入力してください',
    passwordMismatch: '2つのパスワードが一致しません', termsRequired: '先に利用規約とプライバシーポリシーに同意してください', captchaLoadFailed: '認証コードの読み込みに失敗しました',
    loginFailed: 'ログインに失敗しました', registerFailed: '登録に失敗しました', registerSuccess: '登録が完了しました', resetEmailFailed: '再設定メールの送信に失敗しました',
  },
  server: { passwordResetMailQueued: 'このメールが登録済みの場合、パスワード再設定メールが届きます', passwordResetSuccess: 'パスワードの再設定が完了しました', passwordResetFailed: 'パスワード再設定に失敗しました' },
}

const it: AuthMessages = {
  locale: { label: 'Italiano', short: 'IT' }, login: 'Accedi', register: 'Registrati',
  loginTitle: 'Accedi al tuo account', registerTitle: 'Crea un account', forgotTitle: 'Reimposta password',
  loginSubtitle: 'Bentornato. Continua le tue discussioni e i tuoi scritti.', registerSubtitle: 'Unisciti a GooseForum e inizia il tuo spazio di discussione.',
  forgotSubtitle: 'Inserisci la tua email e ti invieremo un messaggio per reimpostare la password.', usernameOrEmail: 'Nome utente o email', username: 'Nome utente',
  email: 'Email', registeredEmail: 'Email registrata', password: 'Password', newPassword: 'Nuova password', confirmPassword: 'Conferma password', captcha: 'Captcha',
  captchaAlt: 'Captcha', refreshCaptcha: 'Aggiorna captcha', forgotPassword: 'Password dimenticata?',
  agreeTerms: 'Ho letto e accetto i termini e l’informativa sulla privacy', createAccount: 'Crea account', sendResetEmail: 'Invia email di reimpostazione',
  backToLogin: 'Torna all’accesso', continueWith: 'Oppure continua con',
  resetPasswordTitle: 'Reimposta password', resetPasswordSubtitle: 'Imposta una nuova password di accesso. Dopo l’invio, torna ad accedere con essa.',
  resetMissingToken: 'Nel link di reimpostazione manca un token. Riaprilo dalla tua email.', passwordMinLength: 'La password deve contenere almeno 6 caratteri', saveNewPassword: 'Salva nuova password',
  passwordAdviceTitle: 'Consigli per la sicurezza della password', passwordAdviceDescription: 'Imposta una nuova password usata solo per GooseForum. I link di reimpostazione hanno validità limitata; richiedi una nuova email dopo la scadenza.',
  passwordAdvice: { length: 'Almeno 6 caratteri', unique: 'Evita di riutilizzare password di altri siti', loginAfterReset: 'Dopo la reimpostazione, torna ad accedere con la nuova password' },
  validation: {
    loginRequired: 'Inserisci account, password e captcha', registerRequired: 'Completa il modulo di registrazione', forgotRequired: 'Inserisci email e captcha',
    passwordMismatch: 'Le due password non coincidono', termsRequired: 'Accetta prima i termini e l’informativa sulla privacy', captchaLoadFailed: 'Caricamento del captcha non riuscito',
    loginFailed: 'Accesso non riuscito', registerFailed: 'Registrazione non riuscita', registerSuccess: 'Registrazione riuscita', resetEmailFailed: 'Invio dell’email di reimpostazione non riuscito',
  },
  server: { passwordResetMailQueued: 'Se questa email è registrata, riceverai un’email per reimpostare la password', passwordResetSuccess: 'Password reimpostata con successo', passwordResetFailed: 'Reimpostazione della password non riuscita' },
}

export const authResources = { zh, en, ja, it } as const

export function normalizeLocale(value?: string | null): Locale | undefined {
  const short = (value || '').trim().toLowerCase().split(/[-_,;]/)[0] as Locale
  return supportedLocales.includes(short) ? short : undefined
}

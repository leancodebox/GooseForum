// Canonical translation data; compatibility exports and React loaders share this file.
import en from "./en-server-messages.js";

export default {
  ...en,
  "page.notFound": "ページが存在しないか、削除されています。",
  "route.notFound":
    "ルートが見つかりません。URL とリクエストメソッドを確認してください。",
  "admin.moderator.userRequired": "Please enter a moderator user.",
  "admin.moderator.userNotFound": "Moderator user not found.",
  "admin.moderator.notFound": "Moderator record not found.",
  "report.targetInvalid": "通報対象が存在しないか、通報できません。",
  "report.ownContent": "自分のコンテンツは通報できません。",
  "report.duplicate": "すでに通報済みで、処理待ちです。",
  "report.createFailed":
    "通報の送信に失敗しました。しばらくしてから再試行してください。",
  "report.notFound": "通報が見つかりません。",
  "upload.dailyLimit":
    "本日すでに {count} 個のファイルをアップロードし、上限に達しました",
  "upload.dailyLimit.avatar":
    "本日すでに {count} 個のファイルをアップロードしています。アバターのアップロードには {fileCount} 枠が必要で、上限を超えます",
  "auth.required": "続行するにはログインしてください。",
  "auth.signupDisabled": "現在、新規登録は無効になっています。",
  "auth.emailDomain.invalid": "メールアドレスの形式が正しくありません。",
  "auth.emailDomain.notAllowed": "このメールドメインは使用できません。",
  "auth.username.invalid": "ユーザー名の形式が正しくありません。",
  "auth.username.exists": "このユーザー名はすでに使用されています。",
  "auth.email.exists": "このメールアドレスはすでに使用されています。",
  "auth.password.tooShort": "パスワードは {minLength} 文字以上にしてください。",
  "auth.password.tooLong": "パスワードが長すぎます。",
  "auth.password.needsLetterNumber": "パスワードには英字と数字の両方が必要です。",
  "auth.captcha.invalid": "認証コードが正しくないか、有効期限が切れています。",
  "auth.register.failed": "登録に失敗しました。しばらくしてから再試行してください。",
  "auth.register.retryLogin": "登録は完了しましたが、自動ログインに失敗しました。手動でログインしてください。",
  "auth.register.emailVerify": "登録が完了しました。メールアドレスを確認してください。",
  "auth.login.success": "ログインしました。",
  "auth.login.invalidRequest": "ログイン要求の有効期限が切れました。ページを更新してください。",
  "auth.password.invalidFormat": "パスワードの形式が正しくありません。",
  "auth.credentials.invalid": "ユーザー名、メールアドレス、またはパスワードが正しくありません。",
  "auth.account.frozen": "このアカウントは停止されています。",
  "auth.email.unverified": "先にメールアドレスを確認してください。",
  "auth.login.failed": "ログインに失敗しました。",
} as const;

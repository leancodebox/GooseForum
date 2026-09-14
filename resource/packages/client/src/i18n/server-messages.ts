import type { Locale } from "./auth.js";

const resources = {
  zh: {
    "page.notFound": "页面不存在，或已经被删除。",
    "route.notFound": "路由未定义，请确认 URL 和请求方法是否正确。",
    "admin.moderator.userRequired": "请输入版主用户。",
    "admin.moderator.userNotFound": "版主用户不存在。",
    "admin.moderator.notFound": "版主记录不存在。",
    "report.targetInvalid": "举报对象不存在或不可举报。",
    "report.ownContent": "不能举报自己的内容。",
    "report.duplicate": "已举报，等待处理。",
    "report.createFailed": "举报提交失败，请稍后重试。",
    "report.notFound": "举报不存在。",
    "upload.dailyLimit": "您今日已上传 {count} 个文件，已达到每日限制",
    "upload.dailyLimit.avatar":
      "您今日已上传 {count} 个文件，上传头像需要 {fileCount} 个名额，已超过每日限制",
  },
  en: {
    "page.notFound": "The page does not exist or has been deleted.",
    "route.notFound":
      "Route not found. Please check the URL and request method.",
    "admin.moderator.userRequired": "Please enter a moderator user.",
    "admin.moderator.userNotFound": "Moderator user not found.",
    "admin.moderator.notFound": "Moderator record not found.",
    "report.targetInvalid":
      "The reported content does not exist or cannot be reported.",
    "report.ownContent": "You cannot report your own content.",
    "report.duplicate": "Already reported and waiting for review.",
    "report.createFailed": "Failed to submit report. Please try again later.",
    "report.notFound": "Report not found.",
    "upload.dailyLimit":
      "You have uploaded {count} files today and reached the daily limit",
    "upload.dailyLimit.avatar":
      "You have uploaded {count} files today. Uploading an avatar needs {fileCount} slots and would exceed the daily limit",
  },
  ja: {
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
  },
  it: {
    "page.notFound": "La pagina non esiste o è stata eliminata.",
    "route.notFound":
      "Route non trovata. Controlla l’URL e il metodo della richiesta.",
    "admin.moderator.userRequired": "Inserisci un utente moderatore.",
    "admin.moderator.userNotFound": "Utente moderatore non trovato.",
    "admin.moderator.notFound": "Record del moderatore non trovato.",
    "report.targetInvalid":
      "Il contenuto segnalato non esiste o non può essere segnalato.",
    "report.ownContent": "Non puoi segnalare i tuoi contenuti.",
    "report.duplicate": "Già segnalato e in attesa di revisione.",
    "report.createFailed":
      "Invio della segnalazione non riuscito. Riprova più tardi.",
    "report.notFound": "Segnalazione non trovata.",
    "upload.dailyLimit":
      "Hai caricato {count} file oggi e hai raggiunto il limite giornaliero",
    "upload.dailyLimit.avatar":
      "Hai caricato {count} file oggi. Caricare un avatar richiede {fileCount} slot e supererebbe il limite giornaliero",
  },
} as const;

export const serverMessageResources = resources satisfies Record<
  Locale,
  Record<string, string>
>;

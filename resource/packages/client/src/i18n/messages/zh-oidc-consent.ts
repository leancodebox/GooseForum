// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "使用 {site} 登录",
  subtitle: "授权请求",
  verifying: "正在验证授权请求…",
  expired: "授权请求已失效，请返回应用重新发起登录。",
  loadFailed: "无法加载授权请求。",
  decisionFailed: "授权请求处理失败。",
  back: "返回 {site}",
  accessAccount: "希望访问你的 {site} 账号",
  permissions: "允许后，此应用可以：",
  clientId: "客户端 ID",
  trust: "请确认你信任此应用。你可以随时在账号设置中撤销访问权限。",
  deny: "拒绝",
  approve: "允许并继续",
  loading: "处理中…",
  scopes: {
    openid: "确认你的身份",
    profile: "读取名称、用户名和头像",
    email: "读取邮箱和验证状态",
    offline_access: "在你离开后继续保持登录",
  },
} as const;

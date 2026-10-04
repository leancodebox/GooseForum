export default {
  title: '双因素认证', enabled: '已启用', disabled: '未启用',
  cancel: '取消绑定',
  unavailable: '双因素认证暂不可用，请联系管理员。',
  code: '验证器验证码或恢复码', verify: '验证', enable: '绑定验证器', confirm: '启用双因素认证',
  disable: '关闭双因素认证', regenerate: '重新生成恢复码', password: '账号密码',
  failed: '验证失败或已过期，请检查密码和验证码。',
  statusFailed: '双因素认证设置加载失败。', retry: '重试',
  secret: '绑定密钥', qr: '验证器绑定二维码', recoveryTitle: '恢复码',
  recoveryNotice: '请立即保存这些恢复码，每个只能使用一次。所有设备已退出登录。',
  download: '下载恢复码', back: '返回登录', remaining: '剩余 {count} 个恢复码',
  logoutNotice: '修改此设置后，所有设备将退出登录。',
  oauthPassword: '使用第三方登录时，请先在个人资料绑定并验证邮箱，再通过重置密码设置账号密码。',
} as const

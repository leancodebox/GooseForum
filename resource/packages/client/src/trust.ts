export const trustLevels = [0, 100, 200, 1000] as const
export function trustLevelName(level: number, locale = 'en') {
  const names = locale.startsWith('zh') ? ['新用户', '成员', '稳定成员', '社区骨干'] : locale.startsWith('ja') ? ['新規ユーザー', 'メンバー', '常連メンバー', 'コミュニティリーダー'] : locale.startsWith('it') ? ['Nuovo utente', 'Membro', 'Membro abituale', 'Leader della comunità'] : ['New user', 'Member', 'Regular', 'Community leader']
  return names[trustLevels.indexOf(level as typeof trustLevels[number])] || String(level)
}
export interface TrustRule { level: number; enabled: boolean; minAccountAgeHours: number; minVisitDays: number; minTopicsVisited: number }
export interface TrustLimits { level: number; allowImageUpload: boolean; maxDailyTopics: number | null; maxDailyReplies: number | null; maxImagesPerPost: number | null; maxExternalLinksPerPost: number | null; maxMentionsPerPost: number | null }
export interface TrustConfig { schemaVersion: number; autoPromotionEnabled: boolean; promotionRules: TrustRule[]; restrictionsEnabled: boolean; levelLimits: TrustLimits[] }
export interface TrustProgress { level: number; name: string; locked: boolean; version: number; visitDays: number; topicsVisited: number; createdAt: string; next: TrustRule | null; autoPromotionEnabled: boolean; definitions: { level: number; name: string }[] }
export interface TrustHistory { id: number; userId: number; fromLevel: number; toLevel: number; fromLocked: boolean; toLocked: boolean; source: string; operatorId: number; reason: string; createdAt: string }

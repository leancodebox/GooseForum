export const trustLevels = [0, 100, 200, 1000] as const
export interface TrustRule { level: number; enabled: boolean; minAccountAgeHours: number; minVisitDays: number; minTopicsVisited: number }
export interface RegistrationCooldown { minPostAccountAgeMinutes: number; minUploadAccountAgeMinutes: number }
export interface TrustLimits { level: number; allowImageUpload: boolean; allowAttachmentUpload?: boolean; allowProfileCover?: boolean; maxDailyUploads: number | null; maxDailyTopics: number | null; maxDailyReplies: number | null; maxImagesPerPost: number | null; maxExternalLinksPerPost: number | null; maxMentionsPerPost: number | null }
export interface TrustConfig { schemaVersion: number; registrationCooldown: RegistrationCooldown; promotionRules: TrustRule[]; levelLimits: TrustLimits[] }


export interface TrustProgress { level: number; name: string; locked: boolean; version: number; visitDays: number; topicsVisited: number; createdAt: string; next: TrustRule | null; definitions: { level: number; name: string }[]; limits?: TrustLimits; nextLimits?: TrustLimits | null; registrationCooldown?: RegistrationCooldown; communityUploadRoleExempt?: boolean; communityUploadsEnabled?: boolean; restrictionStatus?: string }
export interface TrustHistory { id: number; userId: number; fromLevel: number; toLevel: number; fromLocked: boolean; toLocked: boolean; source: string; operatorId: number; reason: string; createdAt: string }

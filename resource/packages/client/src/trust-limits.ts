export const trustQuotaFields = ['maxDailyTopics', 'maxDailyReplies', 'maxDailyUploads', 'maxImagesPerPost', 'maxExternalLinksPerPost', 'maxMentionsPerPost'] as const
export type TrustLimitField = typeof trustQuotaFields[number] | 'allowImageUpload'

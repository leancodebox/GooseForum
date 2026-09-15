// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  "page.notFound": "The page does not exist or has been deleted.",
  "route.notFound": "Route not found. Please check the URL and request method.",
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
} as const;

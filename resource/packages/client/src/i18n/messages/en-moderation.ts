// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Moderator workspace",
  description: "Handle community management tasks within your scope.",
  notice:
    "With great power comes great responsibility. Check the rules and full context before acting.",
  tabs: {
    reports: "Reports",
    ban: "Blocked posts",
    logs: "Activity log",
    guidance: "Reminders",
  },
  reports: {
    loading: "Loading reports…",
    loadMore: "Load more",
    noMore: "No more reports",
    emptyTitle: "No reports",
    emptyDescription: "User reports will appear here.",
    statusTabs: {
      open: "Pending",
      closed: "Handled",
    },
    targetTypes: {
      topic: "Topic",
      post: "Post",
    },
    reasons: {
      spam: "Spam",
      abuse: "Harassment",
      illegal: "Illegal content",
      irrelevant: "Off topic",
      other: "Other",
    },
    resolutions: {
      banned: "Blocked",
      ignored: "Ignored",
      resolved: "Handled",
    },
    reason: "Reason",
    status: "Status",
    reporter: "Reporter",
    handler: "Handler",
    submitted: "Submitted",
    handled: "Handled",
    noExcerpt: "No excerpt",
    block: "Block",
    ignore: "Ignore",
    loadFailed: "Failed to load reports.",
    actionFailed: "Moderation action failed.",
  },
  blocked: {
    emptyTitle: "No blocked posts",
    emptyDescription: "There is nothing to handle in your scope right now.",
    restore: "Restore",
    action: "Action",
    loadFailed: "Failed to update the post.",
    next: "Next page",
  },
  logs: {
    loading: "Loading logs…",
    loadMore: "Load more",
    noMore: "No more logs",
    emptyTitle: "No activity yet",
    emptyDescription: "Blocks, restores, and scope changes will appear here.",
    operation: "Operation",
    time: "Time",
    loadFailed: "Failed to load activity.",
    actions: {
      topicBlocked: "blocked",
      topicUnblocked: "restored",
      replyBlocked: "hid",
      replyUnblocked: "restored",
      postBlocked: "hid",
      postUnblocked: "restored",
      reportResolved: "resolved report",
      reportRejected: "ignored report",
      categoryModeratorAdded: "added moderator scope",
      categoryModeratorRemoved: "removed moderator scope",
      operation: "updated",
    },
  },
  guidance: {
    rule: {
      title: "Check the rules first",
      description:
        "Do not act on disagreement alone. Confirm a clear rules violation.",
    },
    context: {
      title: "Read the full context",
      description:
        "Review the title, body, category, and discussion background.",
    },
    restraint: {
      title: "Use power with restraint",
      description: "Prefer reminders and conversation when they can work.",
    },
  },
} as const;

// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  title: "Moderator workspace",
  description: "Handle community management tasks within your scope.",
  notice:
    "With great power comes great responsibility. Before blocking a post, make sure it violates the community rules and communicate first when appropriate.",
  guidanceItems: {
    rule: {
      title: "Check the rules first",
      description:
        "Do not act on disagreement alone. Start by confirming whether the post clearly violates community rules.",
    },
    context: {
      title: "Read the full context",
      description:
        "Review the title, body, category, and discussion background before making a decision.",
    },
    restraint: {
      title: "Use power with restraint",
      description:
        "Prefer reminders and conversation when they can work. Blocking is for clear and necessary cases.",
    },
  },
  total: "{count} items",
  blocked: "Blocked",
  emptyDescription: "There is nothing to handle in your scope right now.",
  managementTabs: {
    reports: "Reports",
    ban: "Blocked posts",
    logs: "Activity log",
    guidance: "Reminders",
    posts: "Moderation",
    members: "Members",
    settings: "Rules",
  },
  reports: {
    loading: "Loading reports",
    loadMore: "Load more",
    noMore: "No more reports",
    emptyTitle: "No pending reports",
    emptyDescription: "User reports will appear here.",
    statusTabs: {
      open: "Pending",
      closed: "Handled",
    },
    table: {
      report: "Report",
      reason: "Report info",
      time: "Time",
      action: "Action",
    },
    ban: "Block",
    reject: "Ignore",
    resolve: "Resolved",
    noExcerpt: "No excerpt",
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
  },
  logs: {
    title: "Activity log",
    description:
      "Review moderation console actions in time order so decisions can be traced later.",
    loading: "Loading logs",
    loadMore: "Load more",
    noMore: "No more logs",
    emptyTitle: "No activity yet",
    emptyDescription:
      "Blocks, restores, and moderator scope changes will appear here.",
    table: {
      operation: "Operation",
      time: "Time",
    },
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
  tabs: {
    all: "All",
    blocked: "Blocked",
    ban: "All",
    unban: "Blocked",
  },
  table: {
    topic: "Post",
    updatedAt: "Updated",
    author: "Author",
    activity: "Activity",
    action: "Action",
  },
  meta: {
    views: "Views",
    replies: "Replies",
  },
} as const;

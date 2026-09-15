// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  presetsTitle: "Presets",
  presetApplied: "{name} preset applied. Save the draft to keep it.",
  saveDraft: "Save draft",
  publishSite: "Publish site",
  workflowTitle: "How it works:",
  workflowDescription:
    "You are editing a theme draft. Saving the draft does not affect the live site; after publishing, the custom theme takes over only when the enable switch on the left is on.",
  draftSaved: "Theme draft saved. The live site is unchanged.",
  published: "Theme published to the live site.",
  saveFailed: "Failed to save theme draft",
  publishFailed: "Failed to publish theme",
  restoredToSaved: "Restored to the saved configuration",
  themeRestoredDefault: "{name} restored to the built-in default",
  allRestoredDefault:
    "Restored to the built-in default themes. Save the pre-publish draft to publish site-wide.",
  cssCopied: "CSS copied",
  cssCopyFailed: "Failed to copy CSS",
  pageTitle: "Theme preview settings",
  resetDefault: "Default",
  restoreToSavedTitle: "Restore to the saved configuration",
  enableLabel: "Enable",
  tabLatest: "Latest",
  tabHot: "Hot",
  tabFeatured: "Featured",
  samplePublish: "Publish topic",
  sampleTopic1:
    "Theme system refactor discussion: colors, radius, and component states",
  sampleTopic2: "Readability of Markdown body text in dark mode",
  sampleTopic3: "Visual check of new-user onboarding and notifications",
  sampleTopicExcerpt:
    "Observe how muted text, dividers, tags, and hover backgrounds are layered under the current theme.",
  sampleTopicTitle: "The title of a topic post",
  sampleTopicBody:
    "This simulates body text, links, quotes, and code blocks. Theme variables should keep content readable in both light and dark modes — especially body text, borders, and muted text.",
  sampleTopicQuote: "Color hierarchy should be quiet, but never vague.",
  sampleMessageIncoming: "Is this background comfortable?",
  sampleMessageOutgoing: "The contrast needs to stay solid.",
  sampleTextarea: "Theme variables cover inputs, focus, and body text.",
  presets: {
    goose: {
      label: "Goose",
      description: "Current default, clean and steady",
    },
    clean: {
      label: "Clean",
      description: "Product-minded, muted blue and teal",
    },
    warm: {
      label: "Warm",
      description: "Softer warmth for community spaces",
    },
    deep: {
      label: "Deep",
      description: "Black-led with deeper contrast",
    },
    cupcake: {
      label: "Cupcake",
      description: "Soft pink, light and sweet",
    },
    retro: {
      label: "Retro",
      description: "Warm vintage tones for casual communities",
    },
    synthwave: {
      label: "Synthwave",
      description: "Neon purple and cyan with stronger presence",
    },
    aqua: {
      label: "Aqua",
      description: "Fresh cyan with bright contrast",
    },
    forest: {
      label: "Forest",
      description: "Layered green for knowledge communities",
    },
  },
} as const;

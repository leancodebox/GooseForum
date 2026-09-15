// Canonical translation data; compatibility exports and React loaders share this file.
export default {
  moderationPending:
    "Your content was saved and will be public after approval.",
  moderationRejected:
    "Your content was saved but did not pass moderation and is not public. Edit and resubmit it, or contact an administrator.",
  createTitle: "Publish topic",
  editTitle: "Edit topic",
  subtitle:
    "Write a clear title and choose suitable categories so the discussion is easier to find.",
  titlePlaceholder: "Enter topic title",
  maxCategories: "Up to 3",
  mainCategoryHint:
    "The first category, “{category}”, decides who can read this topic",
  mainCategoryBadge: "Main",
  mainCategoryChangeTitle: "Change the main category?",
  mainCategoryChangeDescription:
    "Removing “{current}” makes “{next}” the main category and changes who can read this topic.",
  confirmMainCategoryChange: "Change main category",
  restrictedCategoryHint:
    "Visible only to members who can access this category",
  restrictedCategorySingleHint:
    "A restricted category must be selected alone; selecting it replaces the current categories",
  noCreatePermission:
    "You can no longer publish in the selected categories, but you can still save a draft.",
  markdownMode: "Markdown",
  visualMode: "Editor",
  preview: "Preview",
  pastePlainText: "Paste as plain text",
  clipboardReadFailed:
    "Unable to read the clipboard. Check your browser permissions.",
  visualUnsupported:
    "This body contains task lists. Continue editing it in Markdown mode.",
  processingImage: "Processing image...",
  processingImages: "Processing images {done}/{total}",
  imageInserted: "Image inserted.",
  imagesInserted: "{count} images inserted.",
  moreImageFailures: "; {count} more failed",
  noUploadableImages: "No uploadable image files",
  topicUpdated: "Topic updated.",
  topicPublished: "Topic published.",
  saveFailed: "Failed to save",
  draftSaveFailed: "Failed to save draft",
  saveDraft: "Save draft",
  updateTopic: "Update topic",
  publishTopic: "Publish topic",
  uploadImageTitle:
    "Upload images; multi-select, paste, and drag-and-drop supported",
  bodyPlaceholder:
    "Enter body text, Markdown supported; paste or drag images here",
  visualPlaceholder: "Write and format the body directly",
  dropToUpload: "Release to upload and insert images",
  emptyPreview: "There is no content to preview yet.",
  selectedCategories: "Selected categories",
  leaveTitle: "Save unfinished edits?",
  leaveDescription:
    "Your current content has not been saved. Save it as a draft before leaving so you can continue later.",
  draftRequirement:
    "A draft needs a title, body, and at least one category before it can be saved.",
  continueEditing: "Continue editing",
  leaveWithoutSaving: "Leave without saving",
  fields: {
    title: "Title",
    category: "Category",
    body: "Body",
  },
  validation: {
    requiredFields:
      "Add a title, select at least one category, and complete the body first.",
    categoryRequired: "Select at least one category before publishing.",
  },
  toolbar: {
    bold: "Bold",
    italic: "Italic",
    strike: "Strikethrough",
    inlineCode: "Inline code",
    link: "Link",
    linkUrl: "Enter link URL",
    applyLink: "Apply",
    quote: "Quote",
    code: "Code block",
    bulletList: "Bullet list",
    orderedList: "Ordered list",
    horizontalRule: "Horizontal rule",
    hardBreak: "Hard break",
    table: "Insert table",
    tableSize: "{rows} rows × {columns} columns",
    blockType: "Block type",
    paragraph: "Paragraph",
    heading: "Heading {level}",
    codeBlock: "Code block",
  },
  placeholder: {
    bold: "bold text",
    italic: "italic text",
    strike: "deleted text",
    link: "link text",
    quote: "quoted content",
    listItem: "list item",
  },
  checklist: {
    title: "Publish checklist",
    done: "Filled",
    pending: "Pending",
    characters: "{count} chars",
  },
} as const;

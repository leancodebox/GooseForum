// Each dynamic import is a separate language/namespace boundary.
export const resourceLoaders: Record<
  string,
  () => Promise<{ default: Record<string, any> }>
> = {
  "zh/auth": () => import("@gooseforum/client/i18n/messages/zh-auth"),
  "zh/oidc-consent": () =>
    import("@gooseforum/client/i18n/messages/zh-oidc-consent"),
  "zh/settings": () => import("@gooseforum/client/i18n/messages/zh-settings"),
  "zh/notifications": () =>
    import("@gooseforum/client/i18n/messages/zh-notifications"),
  "zh/messages": () => import("@gooseforum/client/i18n/messages/zh-messages"),
  "zh/moderation": () =>
    import("@gooseforum/client/i18n/messages/zh-moderation"),
  "zh/publish": () => import("@gooseforum/client/i18n/messages/zh-publish"),
  "zh/topic": () => import("@gooseforum/client/i18n/messages/zh-topic"),
  "zh/theme-preview": () =>
    import("@gooseforum/client/i18n/messages/zh-theme-preview"),
  "zh/server-messages": () =>
    import("@gooseforum/client/i18n/messages/zh-server-messages"),
  "zh/site-shell": () =>
    import("@gooseforum/client/i18n/messages/zh-site-shell"),
  "zh/site-links": () =>
    import("@gooseforum/client/i18n/messages/zh-site-links"),
  "zh/site-sponsors": () =>
    import("@gooseforum/client/i18n/messages/zh-site-sponsors"),
  "zh/site-categories": () =>
    import("@gooseforum/client/i18n/messages/zh-site-categories"),
  "zh/site-members": () =>
    import("@gooseforum/client/i18n/messages/zh-site-members"),
  "zh/site-home": () => import("@gooseforum/client/i18n/messages/zh-site-home"),
  "zh/site-search": () =>
    import("@gooseforum/client/i18n/messages/zh-site-search"),
  "zh/site-user": () => import("@gooseforum/client/i18n/messages/zh-site-user"),
  "zh/site-userCard": () =>
    import("@gooseforum/client/i18n/messages/zh-site-userCard"),
  "zh/legacy-site-pages-drafts": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-drafts"),
  "zh/legacy-site-pages-accessGroups": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-accessGroups"),
  "zh/legacy-site-pages-moderation": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-moderation"),
  "zh/legacy-site-pages-publish": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-publish"),
  "zh/legacy-site-pages-topic": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-topic"),
  "zh/legacy-site-pages-themePreview": () =>
    import("@gooseforum/client/i18n/messages/zh-legacy-site-pages-themePreview"),
  "zh/content-error": () =>
    import("@gooseforum/client/i18n/messages/zh-content-error"),
  "zh/content-common": () =>
    import("@gooseforum/client/i18n/messages/zh-content-common"),
  "en/auth": () => import("@gooseforum/client/i18n/messages/en-auth"),
  "en/oidc-consent": () =>
    import("@gooseforum/client/i18n/messages/en-oidc-consent"),
  "en/settings": () => import("@gooseforum/client/i18n/messages/en-settings"),
  "en/notifications": () =>
    import("@gooseforum/client/i18n/messages/en-notifications"),
  "en/messages": () => import("@gooseforum/client/i18n/messages/en-messages"),
  "en/moderation": () =>
    import("@gooseforum/client/i18n/messages/en-moderation"),
  "en/publish": () => import("@gooseforum/client/i18n/messages/en-publish"),
  "en/topic": () => import("@gooseforum/client/i18n/messages/en-topic"),
  "en/theme-preview": () =>
    import("@gooseforum/client/i18n/messages/en-theme-preview"),
  "en/server-messages": () =>
    import("@gooseforum/client/i18n/messages/en-server-messages"),
  "en/site-shell": () =>
    import("@gooseforum/client/i18n/messages/en-site-shell"),
  "en/site-links": () =>
    import("@gooseforum/client/i18n/messages/en-site-links"),
  "en/site-sponsors": () =>
    import("@gooseforum/client/i18n/messages/en-site-sponsors"),
  "en/site-categories": () =>
    import("@gooseforum/client/i18n/messages/en-site-categories"),
  "en/site-members": () =>
    import("@gooseforum/client/i18n/messages/en-site-members"),
  "en/site-home": () => import("@gooseforum/client/i18n/messages/en-site-home"),
  "en/site-search": () =>
    import("@gooseforum/client/i18n/messages/en-site-search"),
  "en/site-user": () => import("@gooseforum/client/i18n/messages/en-site-user"),
  "en/site-userCard": () =>
    import("@gooseforum/client/i18n/messages/en-site-userCard"),
  "en/legacy-site-pages-drafts": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-drafts"),
  "en/legacy-site-pages-accessGroups": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-accessGroups"),
  "en/legacy-site-pages-moderation": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-moderation"),
  "en/legacy-site-pages-publish": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-publish"),
  "en/legacy-site-pages-topic": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-topic"),
  "en/legacy-site-pages-themePreview": () =>
    import("@gooseforum/client/i18n/messages/en-legacy-site-pages-themePreview"),
  "en/content-error": () =>
    import("@gooseforum/client/i18n/messages/en-content-error"),
  "en/content-common": () =>
    import("@gooseforum/client/i18n/messages/en-content-common"),
  "ja/auth": () => import("@gooseforum/client/i18n/messages/ja-auth"),
  "ja/oidc-consent": () =>
    import("@gooseforum/client/i18n/messages/ja-oidc-consent"),
  "ja/settings": () => import("@gooseforum/client/i18n/messages/ja-settings"),
  "ja/notifications": () =>
    import("@gooseforum/client/i18n/messages/ja-notifications"),
  "ja/messages": () => import("@gooseforum/client/i18n/messages/ja-messages"),
  "ja/moderation": () =>
    import("@gooseforum/client/i18n/messages/ja-moderation"),
  "ja/publish": () => import("@gooseforum/client/i18n/messages/ja-publish"),
  "ja/topic": () => import("@gooseforum/client/i18n/messages/ja-topic"),
  "ja/theme-preview": () =>
    import("@gooseforum/client/i18n/messages/ja-theme-preview"),
  "ja/server-messages": () =>
    import("@gooseforum/client/i18n/messages/ja-server-messages"),
  "ja/site-shell": () =>
    import("@gooseforum/client/i18n/messages/ja-site-shell"),
  "ja/site-links": () =>
    import("@gooseforum/client/i18n/messages/ja-site-links"),
  "ja/site-sponsors": () =>
    import("@gooseforum/client/i18n/messages/ja-site-sponsors"),
  "ja/site-categories": () =>
    import("@gooseforum/client/i18n/messages/ja-site-categories"),
  "ja/site-members": () =>
    import("@gooseforum/client/i18n/messages/ja-site-members"),
  "ja/site-home": () => import("@gooseforum/client/i18n/messages/ja-site-home"),
  "ja/site-search": () =>
    import("@gooseforum/client/i18n/messages/ja-site-search"),
  "ja/site-user": () => import("@gooseforum/client/i18n/messages/ja-site-user"),
  "ja/site-userCard": () =>
    import("@gooseforum/client/i18n/messages/ja-site-userCard"),
  "ja/legacy-site-pages-drafts": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-drafts"),
  "ja/legacy-site-pages-accessGroups": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-accessGroups"),
  "ja/legacy-site-pages-moderation": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-moderation"),
  "ja/legacy-site-pages-publish": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-publish"),
  "ja/legacy-site-pages-topic": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-topic"),
  "ja/legacy-site-pages-themePreview": () =>
    import("@gooseforum/client/i18n/messages/ja-legacy-site-pages-themePreview"),
  "ja/content-error": () =>
    import("@gooseforum/client/i18n/messages/ja-content-error"),
  "ja/content-common": () =>
    import("@gooseforum/client/i18n/messages/ja-content-common"),
  "it/auth": () => import("@gooseforum/client/i18n/messages/it-auth"),
  "it/oidc-consent": () =>
    import("@gooseforum/client/i18n/messages/it-oidc-consent"),
  "it/settings": () => import("@gooseforum/client/i18n/messages/it-settings"),
  "it/notifications": () =>
    import("@gooseforum/client/i18n/messages/it-notifications"),
  "it/messages": () => import("@gooseforum/client/i18n/messages/it-messages"),
  "it/moderation": () =>
    import("@gooseforum/client/i18n/messages/it-moderation"),
  "it/publish": () => import("@gooseforum/client/i18n/messages/it-publish"),
  "it/topic": () => import("@gooseforum/client/i18n/messages/it-topic"),
  "it/theme-preview": () =>
    import("@gooseforum/client/i18n/messages/it-theme-preview"),
  "it/server-messages": () =>
    import("@gooseforum/client/i18n/messages/it-server-messages"),
  "it/site-shell": () =>
    import("@gooseforum/client/i18n/messages/it-site-shell"),
  "it/site-links": () =>
    import("@gooseforum/client/i18n/messages/it-site-links"),
  "it/site-sponsors": () =>
    import("@gooseforum/client/i18n/messages/it-site-sponsors"),
  "it/site-categories": () =>
    import("@gooseforum/client/i18n/messages/it-site-categories"),
  "it/site-members": () =>
    import("@gooseforum/client/i18n/messages/it-site-members"),
  "it/site-home": () => import("@gooseforum/client/i18n/messages/it-site-home"),
  "it/site-search": () =>
    import("@gooseforum/client/i18n/messages/it-site-search"),
  "it/site-user": () => import("@gooseforum/client/i18n/messages/it-site-user"),
  "it/site-userCard": () =>
    import("@gooseforum/client/i18n/messages/it-site-userCard"),
  "it/legacy-site-pages-drafts": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-drafts"),
  "it/legacy-site-pages-accessGroups": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-accessGroups"),
  "it/legacy-site-pages-moderation": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-moderation"),
  "it/legacy-site-pages-publish": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-publish"),
  "it/legacy-site-pages-topic": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-topic"),
  "it/legacy-site-pages-themePreview": () =>
    import("@gooseforum/client/i18n/messages/it-legacy-site-pages-themePreview"),
  "it/content-error": () =>
    import("@gooseforum/client/i18n/messages/it-content-error"),
  "it/content-common": () =>
    import("@gooseforum/client/i18n/messages/it-content-common"),
};

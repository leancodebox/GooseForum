import type { ThemePayload } from "@gooseforum/client";

export function NextThemeResources({ theme }: { theme: ThemePayload }) {
  return (
    <>
      <meta
        name="theme-color"
        content={theme.themeColor}
        data-goose-theme-color=""
      />
      {theme.enabled && theme.href ? (
        <link
          id="goose-site-theme-link"
          rel="stylesheet"
          href={theme.href}
          precedence="goose-theme"
        />
      ) : null}
    </>
  );
}

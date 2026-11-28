import "@gooseforum/markdown/styles.css";
import { use, useMemo } from "react";
import { renderMarkdown } from "@gooseforum/markdown";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";

const guides = import.meta.glob<string>("../docs/http-notify-guide.*.md", {
  query: "?raw",
  import: "default",
});
const guideRequests = new Map<AuthLocale, Promise<string>>();

function loadGuide(locale: AuthLocale) {
  const pending = guideRequests.get(locale);
  if (pending) return pending;
  const load = guides["../docs/http-notify-guide." + locale + ".md"] ??
    guides["../docs/http-notify-guide.en.md"];
  const request = load().catch((error: unknown) => {
    guideRequests.delete(locale);
    throw error;
  });
  guideRequests.set(locale, request);
  return request;
}

export default function HttpNotifyGuide({ locale }: { locale: AuthLocale }) {
  const source = use(loadGuide(locale));
  const html = useMemo(() => renderMarkdown(source), [source]);
  return <article className="typeset min-w-0 max-w-3xl" dangerouslySetInnerHTML={{ __html: html }} />;
}

import { useMemo } from "react";
import { renderMarkdown } from "@gooseforum/markdown";
import type { AuthLocale } from "@gooseforum/runtime/i18n/auth";
import en from "../docs/http-notify-guide.en.md?raw";
import zh from "../docs/http-notify-guide.zh.md?raw";
import ja from "../docs/http-notify-guide.ja.md?raw";

export default function HttpNotifyGuide({ locale }: { locale: AuthLocale }) {
  const source = locale === "zh" ? zh : locale === "ja" ? ja : en;
  const html = useMemo(() => renderMarkdown(source), [source]);
  return <article className="typeset min-w-0 max-w-3xl" dangerouslySetInnerHTML={{ __html: html }} />;
}

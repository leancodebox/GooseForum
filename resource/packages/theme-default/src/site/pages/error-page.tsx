import type { ErrorPageProps } from "@gooseforum/client";
import { ArrowLeft, Home } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Button } from "@gooseforum/ui/components/button";
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink } from "@gooseforum/runtime";

export function ErrorPageView({ page }: { page: ErrorPageProps }) {
  const { t } = useTranslation("error");
  const { t: serverT, i18n } = useTranslation("serverMessages");
  const title = page.code === "404" ? t("notFoundTitle") : page.title;
  const knownMessage =
    page.messageCode && i18n.exists(page.messageCode, { ns: "serverMessages" })
      ? serverT(page.messageCode, page.params)
      : page.messageCode === "page.notFound" ||
          page.messageCode === "route.notFound"
        ? t(`messages.${page.messageCode}`, page.params)
        : "";
  function back() {
    if (window.history.length > 1) window.history.back();
    else window.location.assign("/");
  }
  return (
    <main className="min-w-0 pb-8">
      <section className="site-panel overflow-hidden rounded-xl border bg-background">
        <Empty className="min-h-72 border-0">
          <EmptyHeader>
            <EmptyTitle className="text-lg">
              {page.code} · {title}
            </EmptyTitle>
            <EmptyDescription>
              {knownMessage || t("fallbackMessage")}
            </EmptyDescription>
          </EmptyHeader>
          <EmptyContent className="flex-row justify-center">
            <Button type="button" variant="outline" onClick={back}>
              <ArrowLeft data-icon="inline-start" />
              {t("back")}
            </Button>
            <Button asChild>
              <GooseLink href="/">
                <Home data-icon="inline-start" />
                {t("home")}
              </GooseLink>
            </Button>
          </EmptyContent>
        </Empty>
      </section>
    </main>
  );
}

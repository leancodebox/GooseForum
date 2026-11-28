import { Suspense } from "react";
import {
  ChevronRightIcon,
  LayoutGridIcon,
  MessageCircleIcon,
} from "lucide-react";
import {
  formatCompactNumber,
  type CategoriesPageProps,
} from "@gooseforum/client";
import { useTranslation } from "react-i18next";
import { Badge } from "@gooseforum/ui/components/badge";
import { CategoryIcon } from "@gooseforum/ui/components/category-icon";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@gooseforum/ui/components/empty";
import { GooseLink } from "@gooseforum/runtime";
import { PageHeader } from "../layout/page-header";

export function CategoriesPageView({ page }: { page: CategoriesPageProps }) {
  const { t } = useTranslation("site.categories");
  return (
    <div className="pb-12">
      <PageHeader
        compact
        title={t("title")}
        description={t("subtitle")}
        badge={
          <Badge variant="secondary">{t("total", { count: page.total })}</Badge>
        }
      />
      {page.categories.length ? (
        <section className="grid lg:grid-cols-2 lg:gap-3">
          {page.categories.map((category) => (
            <GooseLink
              key={category.id}
              href={category.url}
              className="group flex min-w-0 items-stretch border-b bg-background p-3.5 transition-colors hover:border-primary/30 hover:bg-muted lg:rounded-xl lg:border"
            >
              <span
                className="mr-3 w-1 shrink-0 rounded-full"
                style={{ backgroundColor: category.color || "var(--primary)" }}
                aria-hidden="true"
              />
              <div className="min-w-0 flex-1 py-0.5">
                <Suspense fallback={<div aria-hidden="true" className="h-[22.5px]" />}>
                  <div className="flex min-w-0 items-center gap-2">
                    <CategoryIcon icon={category.icon} color={category.color || undefined} suspend />
                    <h2 className="truncate text-[15px] font-bold transition-colors group-hover:text-primary">
                      {category.name}
                    </h2>
                  </div>
                </Suspense>
                <p className="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                  {category.description || t("noDescription")}
                </p>
              </div>
              <div className="ml-3 flex shrink-0 items-center gap-2 self-center">
                <Badge variant="secondary" className="gap-1 text-[10px]">
                  <MessageCircleIcon data-icon="inline-start" />
                  {t("topicCount", {
                    total: formatCompactNumber(category.topicCount),
                  })}
                </Badge>
                <ChevronRightIcon
                  width={16}
                  height={16}
                  className="text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary"
                  aria-hidden="true"
                />
              </div>
            </GooseLink>
          ))}
        </section>
      ) : (
        <Empty className="min-h-56 border bg-background">
          <EmptyHeader>
            <EmptyMedia variant="icon">
              <LayoutGridIcon />
            </EmptyMedia>
            <EmptyTitle>{t("emptyTitle")}</EmptyTitle>
            <EmptyDescription>{t("emptyDescription")}</EmptyDescription>
          </EmptyHeader>
        </Empty>
      )}
    </div>
  );
}

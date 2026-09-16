"use client";

import { memo, Suspense } from "react";
import { preparedPage } from "@gooseforum/runtime/prepared-page";
import type { AnyPagePayload } from "@gooseforum/client";
import { ArrowRightIcon } from "lucide-react";
import { GooseLink, useGooseRuntime } from "@gooseforum/runtime";
import { Badge } from "@gooseforum/ui/components/badge";
import { Button } from "@gooseforum/ui/components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@gooseforum/ui/components/card";
import { Separator } from "@gooseforum/ui/components/separator";
import { AppShell } from "../site/layout/app-shell";

const LoginPageView = preparedPage("auth.login", () =>
  import("../site/auth/login-page").then((module) => ({
    default: module.LoginPageView,
  })),
);
const ResetPasswordPageView = preparedPage("auth.resetPassword", () =>
  import("../site/auth/reset-password-page").then((module) => ({
    default: module.ResetPasswordPageView,
  })),
);
const OIDCConsentPageView = preparedPage("auth.oidcConsent", () =>
  import("../site/auth/oidc-consent-page").then((module) => ({
    default: module.OIDCConsentPageView,
  })),
);
const LinksPageView = preparedPage("links.index", () =>
  import("../site/pages/links-page").then((module) => ({
    default: module.LinksPageView,
  })),
);
const SponsorsPageView = preparedPage("sponsors.index", () =>
  import("../site/pages/sponsors-page").then((module) => ({
    default: module.SponsorsPageView,
  })),
);
const CategoriesPageView = preparedPage("categories.index", () =>
  import("../site/pages/categories-page").then((module) => ({
    default: module.CategoriesPageView,
  })),
);
const MembersPageView = preparedPage("members.index", () =>
  import("../site/pages/members-page").then((module) => ({
    default: module.MembersPageView,
  })),
);
const HomePageView = preparedPage("home.index", () =>
  import("../site/pages/home-page").then((module) => ({
    default: module.HomePageView,
  })),
);
const CategoryPageView = preparedPage("category.index", () =>
  import("../site/pages/category-page").then((module) => ({
    default: module.CategoryPageView,
  })),
);
const SearchPageView = preparedPage("search.index", () =>
  import("../site/pages/search-page").then((module) => ({
    default: module.SearchPageView,
  })),
);
const UserProfilePageView = preparedPage("user.profile", () =>
  import("../site/pages/user-profile-page").then((module) => ({
    default: module.UserProfilePageView,
  })),
);

const SettingsPageView = preparedPage("settings.index", () =>
  import("../site/settings/settings-page").then((module) => ({
    default: module.SettingsPageView,
  })),
);
const NotificationsPageView = preparedPage("notifications.index", () =>
  import("../site/pages/notifications-page").then((module) => ({
    default: module.NotificationsPageView,
  })),
);
const MessagesPageView = preparedPage("messages.index", () =>
  import("../site/pages/messages-page").then((module) => ({
    default: module.MessagesPageView,
  })),
);
const DraftsPageView = preparedPage("drafts.index", () =>
  import("../site/pages/drafts-page").then((module) => ({
    default: module.DraftsPageView,
  })),
);
const AccessGroupsPageView = preparedPage("access-groups.index", () =>
  import("../site/pages/access-groups-page").then((module) => ({
    default: module.AccessGroupsPageView,
  })),
);
const ErrorPageView = preparedPage("error.index", () =>
  import("../site/pages/error-page").then((module) => ({
    default: module.ErrorPageView,
  })),
);
const ModerationPageView = preparedPage("moderation.index", () =>
  import("../site/pages/moderation-page").then((module) => ({
    default: module.ModerationPageView,
  })),
);
const PublishPageView = preparedPage("publish.index", () =>
  import("../site/pages/publish-page").then((module) => ({
    default: module.PublishPageView,
  })),
);
const TopicPageView = preparedPage("topic.detail", () =>
  import("../site/pages/topic-page").then((module) => ({
    default: module.TopicPageView,
  })),
);
const ThemePreviewPageView = preparedPage("theme.preview", () =>
  import("../site/pages/theme-preview-page").then((module) => ({
    default: module.ThemePreviewPageView,
  })),
);

export interface GooseAppProps {
  page: AnyPagePayload;
}

export function GooseApp({ page }: GooseAppProps) {
  const content = <GoosePage page={page} />;
  return isStandalonePage(page) ? (
    content
  ) : (
    <AppShell layout={page.layout}>{content}</AppShell>
  );
}

export const GoosePage = memo(function GoosePage({ page }: GooseAppProps) {
  return (
    <Suspense fallback={<PageLoading />}>
      <GoosePageContent page={page} />
    </Suspense>
  );
});

function GoosePageContent({ page }: GooseAppProps) {
  if (page.component === "auth.login") {
    return <LoginPageView layout={page.layout} page={page.props} />;
  }
  if (page.component === "auth.resetPassword") {
    return <ResetPasswordPageView layout={page.layout} page={page.props} />;
  }
  if (page.component === "auth.oidcConsent") {
    return <OIDCConsentPageView layout={page.layout} page={page.props} />;
  }

  return page.component === "links.index" ? (
        <LinksPageView page={page.props} />
      ) : page.component === "sponsors.index" ? (
        <SponsorsPageView page={page.props} />
      ) : page.component === "categories.index" ? (
        <CategoriesPageView page={page.props} />
      ) : page.component === "members.index" ? (
        <MembersPageView page={page.props} />
      ) : page.component === "home.index" ? (
        <HomePageView
          layout={page.layout}
          page={page.props}
          pageUrl={page.url}
        />
      ) : page.component === "category.index" ? (
        <CategoryPageView page={page.props} pageUrl={page.url} />
      ) : page.component === "search.index" ? (
        <SearchPageView page={page.props} />
      ) : page.component === "user.profile" ? (
        <UserProfilePageView key={page.url} page={page.props} />
      ) : page.component === "settings.index" ? (
        <Suspense
          fallback={
            <div className="grid min-h-48 place-items-center text-sm text-muted-foreground">
              加载中…
            </div>
          }
        >
          <SettingsPageView
            key={page.url}
            layout={page.layout}
            page={page.props}
          />
        </Suspense>
      ) : page.component === "notifications.index" ? (
        <Suspense
          fallback={
            <div className="grid min-h-48 place-items-center text-sm text-muted-foreground">
              加载中…
            </div>
          }
        >
          <NotificationsPageView key={page.url} page={page.props} />
        </Suspense>
      ) : page.component === "messages.index" ? (
        <Suspense
          fallback={
            <div className="grid min-h-48 place-items-center text-sm text-muted-foreground">
              加载中…
            </div>
          }
        >
          <MessagesPageView
            key={page.url}
            layout={page.layout}
            page={page.props}
          />
        </Suspense>
      ) : page.component === "drafts.index" ? (
        <Suspense fallback={<PageLoading />}>
          <DraftsPageView page={page.props} />
        </Suspense>
      ) : page.component === "access-groups.index" ? (
        <Suspense fallback={<PageLoading />}>
          <AccessGroupsPageView />
        </Suspense>
      ) : page.component === "error.index" ? (
        <Suspense fallback={<PageLoading />}>
          <ErrorPageView page={page.props} />
        </Suspense>
      ) : page.component === "moderation.index" ? (
        <Suspense fallback={<PageLoading />}>
          <ModerationPageView page={page.props} />
        </Suspense>
      ) : page.component === "publish.index" ? (
        <Suspense fallback={<PageLoading />}>
          <PublishPageView key={page.url} page={page.props} />
        </Suspense>
      ) : page.component === "topic.detail" ? (
        <Suspense fallback={<PageLoading />}>
          <TopicPageView
            key={page.url}
            layout={page.layout}
            page={page.props}
          />
        </Suspense>
      ) : page.component === "theme.preview" ? (
        <Suspense fallback={<PageLoading />}>
          <ThemePreviewPageView layout={page.layout} page={page.props} />
        </Suspense>
      ) : (
        <PayloadPreview page={page} />
      );
}

export function isStandalonePage(page: AnyPagePayload) {
  return (
    page.component === "auth.login" ||
    page.component === "auth.resetPassword" ||
    page.component === "auth.oidcConsent"
  );
}

function PageLoading() {
  return (
    <div className="grid min-h-48 place-items-center text-sm text-muted-foreground">
      加载中…
    </div>
  );
}

function PayloadPreview({ page }: GooseAppProps) {
  const { component, layout, meta, url, version } = page;
  const { isNavigating, refresh } = useGooseRuntime();
  const nextUrl = component === "auth.login" ? "/" : "/login";
  const nextLabel =
    component === "auth.login" ? "返回首页 payload" : "加载登录页 payload";

  return (
    <main className="mx-auto flex min-h-svh max-w-5xl items-center px-4 py-10 sm:px-6">
      <Card className="w-full">
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <Badge>React</Badge>
            <Badge variant="outline">shadcn/ui</Badge>
            <Badge variant="secondary">payload {version}</Badge>
          </div>
          <CardTitle className="text-xl">
            {layout.site.name || "GooseForum"}
          </CardTitle>
          <CardDescription>
            React 独立开发入口已通过接口连接到 GooseForum 页面 payload。
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4">
          <Separator />
          <dl className="grid gap-3 text-sm sm:grid-cols-2">
            <PayloadField label="页面组件" value={component} />
            <PayloadField label="页面地址" value={url} />
            <PayloadField label="页面标题" value={meta.title} />
            <PayloadField label="当前主题" value={layout.theme.current} />
          </dl>
        </CardContent>
        <CardFooter className="flex-wrap justify-end">
          <Button
            variant="outline"
            disabled={isNavigating}
            onClick={() => void refresh()}
          >
            {isNavigating ? "正在加载…" : "刷新接口数据"}
          </Button>
          <Button asChild>
            <GooseLink href={nextUrl}>
              {nextLabel}
              <ArrowRightIcon data-icon="inline-end" />
            </GooseLink>
          </Button>
        </CardFooter>
      </Card>
    </main>
  );
}

function PayloadField({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1 rounded-lg bg-muted px-3 py-2">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="min-w-0 truncate font-mono">{value || "—"}</dd>
    </div>
  );
}

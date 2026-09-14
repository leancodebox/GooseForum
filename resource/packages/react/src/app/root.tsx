import { lazy, Suspense } from "react";
import type { AnyPagePayload } from "@gooseforum/client";
import { ArrowRightIcon } from "lucide-react";
import { GooseLink, useGooseRuntime } from "../runtime";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Separator } from "../components/ui/separator";
import { Skeleton } from "../components/ui/skeleton";
import { LoginPageView } from "../site/auth/login-page";
import { ResetPasswordPageView } from "../site/auth/reset-password-page";
import { OIDCConsentPageView } from "../site/auth/oidc-consent-page";
import { AppShell } from "../site/layout/app-shell";
import { LinksPageView } from "../site/pages/links-page";
import { SponsorsPageView } from "../site/pages/sponsors-page";
import { CategoriesPageView } from "../site/pages/categories-page";
import { MembersPageView } from "../site/pages/members-page";
import { HomePageView } from "../site/pages/home-page";
import { CategoryPageView } from "../site/pages/category-page";
import { SearchPageView } from "../site/pages/search-page";
import { UserProfilePageView } from "../site/pages/user-profile-page";

const SettingsPageView = lazy(() =>
  import("../site/settings/settings-page").then((module) => ({
    default: module.SettingsPageView,
  })),
);
const NotificationsPageView = lazy(() =>
  import("../site/pages/notifications-page").then((module) => ({
    default: module.NotificationsPageView,
  })),
);

export interface GooseAppProps {
  page: AnyPagePayload;
}

export function GooseApp({ page }: GooseAppProps) {
  if (page.component === "auth.login") {
    return <LoginPageView layout={page.layout} page={page.props} />;
  }
  if (page.component === "auth.resetPassword") {
    return <ResetPasswordPageView layout={page.layout} page={page.props} />;
  }
  if (page.component === "auth.oidcConsent") {
    return <OIDCConsentPageView layout={page.layout} page={page.props} />;
  }

  return (
    <AppShell layout={page.layout}>
      {page.component === "links.index" ? (
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
      ) : (
        <PayloadPreview page={page} />
      )}
    </AppShell>
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

export function BootstrapLoading() {
  return (
    <main
      className="mx-auto flex min-h-svh max-w-2xl items-center px-4 py-10 sm:px-6"
      aria-busy="true"
    >
      <Card className="w-full" role="status">
        <CardHeader>
          <CardTitle>正在读取 GooseForum payload</CardTitle>
          <CardDescription>
            React 已启动，正在通过开发接口获取页面数据。
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          <Skeleton className="h-12 w-full" />
          <Skeleton className="h-12 w-full" />
        </CardContent>
      </Card>
    </main>
  );
}

export function BootstrapError({
  error,
  onRetry,
  retrying = false,
}: {
  error: unknown;
  onRetry?: () => void;
  retrying?: boolean;
}) {
  const message =
    error instanceof Error ? error.message : "Unknown bootstrap error";

  return (
    <main className="mx-auto flex min-h-svh max-w-2xl items-center px-4 py-10 sm:px-6">
      <Card className="w-full border-destructive/40">
        <CardHeader>
          <Badge variant="destructive" className="mb-2">
            连接失败
          </Badge>
          <CardTitle>无法读取 GooseForum payload</CardTitle>
          <CardDescription>
            请先启动 Go 服务，或检查 Vite 的 GOOSE_DEV_ORIGIN 配置。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <pre className="overflow-auto rounded-lg bg-muted p-3 text-xs">
            {message}
          </pre>
        </CardContent>
        {onRetry ? (
          <CardFooter className="justify-end">
            <Button disabled={retrying} onClick={onRetry}>
              {retrying ? "正在重试…" : "重新请求 payload"}
            </Button>
          </CardFooter>
        ) : null}
      </Card>
    </main>
  );
}

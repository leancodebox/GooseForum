import { Badge } from "./components/badge";
import { Button } from "./components/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./components/card";
import { Skeleton } from "./components/skeleton";

export function BootstrapLoading() {
  return (
    <main
      className="mx-auto flex min-h-svh max-w-2xl items-center px-4 py-10 sm:px-6"
      aria-busy="true"
    >
      <Card className="w-full" role="status">
        <CardHeader>
          <CardTitle>正在加载 GooseForum 页面</CardTitle>
          <CardDescription>请稍候，正在准备页面内容。</CardDescription>
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
  description = "请稍后重试；如果问题持续存在，请联系站点管理员。",
}: BootstrapErrorProps) {
  const message =
    error instanceof Error ? error.message : "Unknown bootstrap error";

  return (
    <main className="mx-auto flex min-h-svh max-w-2xl items-center px-4 py-10 sm:px-6">
      <Card className="w-full border-destructive/40">
        <CardHeader>
          <Badge variant="destructive" className="mb-2">
            加载失败
          </Badge>
          <CardTitle>无法加载 GooseForum 页面</CardTitle>
          <CardDescription>{description}</CardDescription>
        </CardHeader>
        <CardContent>
          <pre className="overflow-auto rounded-lg bg-muted p-3 text-xs">
            {message}
          </pre>
        </CardContent>
        {onRetry ? (
          <CardFooter className="justify-end">
            <Button disabled={retrying} onClick={onRetry}>
              {retrying ? "正在重试…" : "重新加载页面"}
            </Button>
          </CardFooter>
        ) : null}
      </Card>
    </main>
  );
}

export interface BootstrapErrorProps {
  error: unknown;
  onRetry?: () => void;
  retrying?: boolean;
  description?: string;
}

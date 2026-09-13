import type { AnyPagePayload } from '@gooseforum/client'
import { ArrowRightIcon } from 'lucide-react'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '../components/ui/card'
import { Separator } from '../components/ui/separator'

export interface GooseAppProps {
  initialPage: AnyPagePayload
}

export function GooseApp({ initialPage }: GooseAppProps) {
  const { component, layout, meta, url, version } = initialPage

  return (
    <main className="mx-auto flex min-h-svh max-w-5xl items-center px-4 py-10 sm:px-6">
      <Card className="w-full">
        <CardHeader>
          <div className="flex flex-wrap items-center gap-2">
            <Badge>React</Badge>
            <Badge variant="outline">shadcn/ui</Badge>
            <Badge variant="secondary">payload {version}</Badge>
          </div>
          <CardTitle className="text-xl">{layout.site.name || 'GooseForum'}</CardTitle>
          <CardDescription>
            React 旁路开发环境已连接到 GooseForum 页面 payload。
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
        <CardFooter className="justify-end">
          <Button asChild>
            <a href={url || '/'}>
              在 Go 版本中打开
              <ArrowRightIcon data-icon="inline-end" />
            </a>
          </Button>
        </CardFooter>
      </Card>
    </main>
  )
}

function PayloadField({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1 rounded-lg bg-muted px-3 py-2">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="min-w-0 truncate font-mono">{value || '—'}</dd>
    </div>
  )
}

export function BootstrapError({ error }: { error: unknown }) {
  const message = error instanceof Error ? error.message : 'Unknown bootstrap error'

  return (
    <main className="mx-auto flex min-h-svh max-w-2xl items-center px-4 py-10 sm:px-6">
      <Card className="w-full border-destructive/40">
        <CardHeader>
          <Badge variant="destructive" className="mb-2">连接失败</Badge>
          <CardTitle>无法读取 GooseForum payload</CardTitle>
          <CardDescription>
            请先启动 Go 服务，或检查 Vite 的 GOOSE_DEV_ORIGIN 配置。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <pre className="overflow-auto rounded-lg bg-muted p-3 text-xs">{message}</pre>
        </CardContent>
      </Card>
    </main>
  )
}

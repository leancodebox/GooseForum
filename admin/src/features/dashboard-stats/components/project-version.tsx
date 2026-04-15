import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import {
  Tag,
  Code2,
  ExternalLink,
  ChevronRight,
  Loader2,
  AlertCircle
} from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'

interface Release {
  id: number
  tag_name: string
  published_at: string
  body: string
  html_url: string
  prerelease: boolean
  draft: boolean
}

interface ProjectVersionProps {
  releases: Release[]
  loading: boolean
  error?: string | null
}

export function ProjectVersion({ releases, loading, error }: ProjectVersionProps) {
  const currentVersion = releases.length > 0 ? releases[0].tag_name : 'v1.0.0'
  const lastUpdate = releases.length > 0 ? releases[0].published_at : null

  return (
    <Card className='flex flex-col h-full'>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-1'>
        <div className='space-y-1'>
          <CardTitle className='flex items-center gap-2 text-lg font-semibold'>
            <Code2 className='h-5 w-5' />
            项目版本
          </CardTitle>
          <CardDescription>系统更新与发布记录</CardDescription>
        </div>
        <Button variant='ghost' size='sm' asChild>
          <a
            href='https://github.com/leancodebox/GooseForum/releases'
            target='_blank'
            rel='noopener noreferrer'
            className='flex items-center gap-1'
          >
            查看全部 <ExternalLink className='h-3 w-3' />
          </a>
        </Button>
      </CardHeader>
      <CardContent className='flex-1 flex flex-col min-h-0 p-4 pt-0'>
        <div className='flex-1 rounded-lg border bg-muted/30 p-4 overflow-y-auto min-h-0' style={{ maxHeight: '350px' }}>
          {loading ? (
            <div className='flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground'>
              <Loader2 className='h-6 w-6 animate-spin' />
              <p className='text-sm'>加载版本信息中...</p>
            </div>
          ) : error ? (
            <div className='flex h-full w-full flex-col items-center justify-center space-y-2 text-destructive'>
              <AlertCircle className='h-6 w-6' />
              <p className='text-sm'>{error}</p>
            </div>
          ) : releases.length > 0 ? (
            <div className='space-y-4'>
              {releases.map((release, index) => (
                <div key={release.id} className='group'>
                  <div className='flex items-start justify-between gap-4'>
                    <div className='flex-1 space-y-1 min-w-0'>
                      <div className='flex items-center gap-2 flex-wrap'>
                        <Tag className='h-4 w-4 text-primary shrink-0' />
                        <span className='font-medium truncate'>{release.tag_name}</span>
                        {release.prerelease && (
                          <Badge variant='secondary' className='bg-amber-100 text-amber-700 hover:bg-amber-100 dark:bg-amber-900/30 dark:text-amber-400 text-[10px] h-4 px-1 shrink-0'>
                            预发布
                          </Badge>
                        )}
                        {release.draft && (
                          <Badge variant='outline' className='text-[10px] h-4 px-1 shrink-0'>
                            草稿
                          </Badge>
                        )}
                      </div>
                      <p className='text-xs text-muted-foreground'>
                        {formatDistanceToNow(new Date(release.published_at), {
                          addSuffix: true,
                          locale: zhCN,
                        })}
                      </p>
                      <p className='mt-2 line-clamp-2 text-sm text-foreground/80 break-words'>
                        {release.body || '暂无发布说明'}
                      </p>
                    </div>
                    <Button variant='ghost' size='icon' className='h-8 w-8 opacity-0 group-hover:opacity-100 transition-opacity shrink-0' asChild>
                      <a href={release.html_url} target='_blank' rel='noopener noreferrer'>
                        <ChevronRight className='h-4 w-4' />
                      </a>
                    </Button>
                  </div>
                  {index < releases.length - 1 && (
                    <div className='mt-4 border-t' />
                  )}
                </div>
              ))}
            </div>
          ) : (
            <div className='flex h-full w-full flex-col items-center justify-center space-y-2 text-muted-foreground'>
              <Code2 className='h-10 w-10 opacity-20' />
              <p className='text-sm'>暂无发布信息</p>
            </div>
          )}
        </div>

        {!loading && !error && releases.length > 0 && (
          <div className='mt-4 flex flex-col gap-1 text-xs text-muted-foreground'>
            <div className='flex items-center gap-2'>
              <div className='h-2 w-2 rounded-full bg-emerald-500' />
              <span>当前版本：{currentVersion}</span>
            </div>
            {lastUpdate && (
              <div className='ml-4'>
                最后更新：{new Date(lastUpdate).toLocaleDateString('zh-CN')}
              </div>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  )
}

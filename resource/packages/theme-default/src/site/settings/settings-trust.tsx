import { useEffect, useState } from 'react'
import { trustLevelName, type TrustProgress } from '@gooseforum/client'
import { useGooseRuntime } from '@gooseforum/runtime'
import { Button } from '@gooseforum/ui/components/button'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { RefreshCw } from 'lucide-react'

export function TrustProgressSection() {
  const { api, locale } = useGooseRuntime()
  const zh = locale.startsWith('zh')
  const [data, setData] = useState<TrustProgress>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [reload, setReload] = useState(0)
  useEffect(() => {
    let active = true
    void api.users.trustLevel().then(p => { if (active) setData(p) }).catch(e => { if (active) setError(String(e)) }).finally(() => { if (active) setLoading(false) })
    return () => { active = false }
  }, [api, reload])
  return <section className="space-y-3 border-t pt-4">
    <h3 className="text-sm font-semibold">{zh ? '信任等级' : 'Trust level'}</h3>
    {loading ? <Spinner /> : error ? <div role="alert" className="flex items-center gap-2 text-sm">{error}<Button type="button" size="icon-sm" variant="outline" aria-label={zh ? '重试' : 'Retry'} title={zh ? '重试' : 'Retry'} onClick={() => { setError(''); setLoading(true); setReload(v => v + 1) }}><RefreshCw /></Button></div> : data ? <>
      <div className="text-sm font-medium">{trustLevelName(data.level, locale)}</div>
      {data.locked ? <p className="text-xs text-muted-foreground">{zh ? '自动晋级已锁定' : 'Automatic promotion locked'}</p> : !data.autoPromotionEnabled ? <p className="text-xs text-muted-foreground">{zh ? '自动晋级已暂停' : 'Automatic promotion paused'}</p> : data.next ? <dl className="grid gap-3 text-sm sm:grid-cols-3">
        <div><dt className="text-xs text-muted-foreground">{zh ? '下一等级' : 'Next level'}</dt><dd>{trustLevelName(data.next.level, locale)}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{zh ? '访问日（UTC）' : 'Visit days (UTC)'}</dt><dd>{data.visitDays} / {data.next.minVisitDays}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{zh ? '不同主题' : 'Distinct topics'}</dt><dd>{data.topicsVisited} / {data.next.minTopicsVisited}</dd></div>
        <div className="sm:col-span-3"><dt className="text-xs text-muted-foreground">{zh ? '注册时长条件到期' : 'Account age requirement met at'}</dt><dd>{new Date(new Date(data.createdAt).getTime() + data.next.minAccountAgeHours * 3600000).toLocaleString(locale)}</dd></div>
      </dl> : null}
    </> : null}
  </section>
}

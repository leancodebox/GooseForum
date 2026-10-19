import { useEffect, useState } from 'react'
import { trustLevelName, type TrustProgress, type TrustLimits } from '@gooseforum/client'
import { trustQuotaFields } from '@gooseforum/client/trust-limits'
import { useGooseRuntime } from '@gooseforum/runtime'
import { Button } from '@gooseforum/ui/components/button'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { RefreshCw } from 'lucide-react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table'

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
  return <section className="space-y-3 border-t p-4">
    <h3 className="text-sm font-semibold">{zh ? '信任等级' : 'Trust level'}</h3>
    {loading ? <Spinner /> : error ? <div role="alert" className="flex items-center gap-2 text-sm">{error}<Button type="button" size="icon-sm" variant="outline" aria-label={zh ? '重试' : 'Retry'} title={zh ? '重试' : 'Retry'} onClick={() => { setError(''); setLoading(true); setReload(v => v + 1) }}><RefreshCw /></Button></div> : data ? <>
      <div className="text-sm font-medium">{trustLevelName(data.level, locale)}</div>
      {data.locked ? <p className="text-xs text-muted-foreground">{zh ? '自动晋级已锁定' : 'Automatic promotion locked'}</p> : data.next ? <dl className="grid gap-3 text-sm sm:grid-cols-3">
        <div><dt className="text-xs text-muted-foreground">{zh ? '下一等级' : 'Next level'}</dt><dd>{trustLevelName(data.next.level, locale)}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{zh ? '访问日（UTC）' : 'Visit days (UTC)'}</dt><dd>{data.visitDays} / {data.next.minVisitDays}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{zh ? '不同主题' : 'Distinct topics'}</dt><dd>{data.topicsVisited} / {data.next.minTopicsVisited}</dd></div>
        <div className="sm:col-span-3"><dt className="text-xs text-muted-foreground">{zh ? '注册时长条件到期' : 'Account age requirement met at'}</dt><dd>{new Date(new Date(data.createdAt).getTime() + data.next.minAccountAgeHours * 3600000).toLocaleString(locale)}</dd></div>
      </dl> : null}
      {data.limits ? <TrustCapabilities data={data} limits={data.limits} locale={locale} /> : null}
    </> : null}
  </section>
}

function TrustCapabilities({ data, limits, locale }: { data: TrustProgress; limits: TrustLimits; locale: string }) {
  const [viewedAt] = useState(() => Date.now())
  const zh = locale.startsWith('zh')
  const next = !data.locked ? data.nextLimits : null
  const labels = {
    maxDailyTopics: zh ? '每日新建主题（含草稿）' : 'Daily topics (including drafts)',
    maxDailyReplies: zh ? '每日回复' : 'Daily replies',
    maxDailyUploads: zh ? '每日上传文件' : 'Daily uploads',
    maxImagesPerPost: zh ? '每篇图片' : 'Images per post',
    maxExternalLinksPerPost: zh ? '每篇外链' : 'External links per post',
    maxMentionsPerPost: zh ? '每篇提及人数' : 'Mentions per post',
  }
  const quota = (value: number | null) => value === null ? (zh ? '不限' : 'Unlimited') : String(value)
  const upload = (value: TrustLimits) => data.communityUploadsEnabled !== false && value.allowImageUpload ? (zh ? '允许' : 'Allowed') : (zh ? '禁止' : 'Not allowed')
  function cooldown(minutes: number) {
    if (minutes === 0) return zh ? '无需等待' : 'No wait'
    const until = new Date(new Date(data.createdAt).getTime() + minutes * 60000)
    return until.getTime() <= viewedAt ? (zh ? '已结束' : 'Completed') : until.toLocaleString(locale)
  }
  return <div className="space-y-3">
    <h4 className="text-sm font-medium">{zh ? '当前等级权限与额度' : 'Current level permissions and limits'}</h4>
    {data.restrictionStatus && data.restrictionStatus !== 'normal' ? <p className="text-sm text-destructive">{zh ? '账号当前受限，等级权限不能绕过账号限制。' : 'Your account is restricted. Level permissions do not override account restrictions.'}</p> : null}
    <Table className="w-full table-fixed text-xs">
      <TableHeader><TableRow><TableHead className="w-1/2 whitespace-normal">{zh ? '项目' : 'Capability'}</TableHead><TableHead className="whitespace-normal">{trustLevelName(data.level, locale)}</TableHead>{next ? <TableHead className="whitespace-normal">{trustLevelName(next.level, locale)}</TableHead> : null}</TableRow></TableHeader>
      <TableBody>
        <TableRow><TableCell className="whitespace-normal">{zh ? '社区图片上传' : 'Community image uploads'}</TableCell><TableCell>{upload(limits)}</TableCell>{next ? <TableCell>{upload(next) === upload(limits) ? (zh ? '不变' : 'No change') : upload(next)}</TableCell> : null}</TableRow>
        {trustQuotaFields.map(field => <TableRow key={field}><TableCell className="whitespace-normal">{labels[field]}</TableCell><TableCell>{quota(limits[field])}</TableCell>{next ? <TableCell>{next[field] === limits[field] ? (zh ? '不变' : 'No change') : quota(next[field])}</TableCell> : null}</TableRow>)}
      </TableBody>
    </Table>
    {next && trustQuotaFields.every(field => next[field] === limits[field]) && upload(next) === upload(limits) ? <p className="text-xs text-muted-foreground">{zh ? '下一等级的以上额度相同。' : 'The next level has the same limits shown above.'}</p> : null}
    {data.registrationCooldown ? <dl className="grid gap-3 text-xs sm:grid-cols-2"><div><dt className="text-muted-foreground">{zh ? '发言冷静期' : 'Posting cooldown'}</dt><dd>{cooldown(data.registrationCooldown.minPostAccountAgeMinutes)}</dd></div><div><dt className="text-muted-foreground">{zh ? '上传冷静期' : 'Upload cooldown'}</dt><dd>{cooldown(data.registrationCooldown.minUploadAccountAgeMinutes)}</dd></div></dl> : null}
    {data.communityUploadRoleExempt ? <p className="text-xs text-muted-foreground">{zh ? '角色账号的社区图片上传豁免每日上传额度和上传冷静期；头像仍受限制。' : 'Your role exempts community image uploads from the daily upload limit and upload cooldown. Avatars remain subject to both.'}</p> : null}
    <p className="text-xs text-muted-foreground">{zh ? '以上为额度上限，不是今日剩余次数；分类访问和账号状态仍独立决定可执行操作。' : 'These are limits, not remaining daily allowances. Category access and account status still apply.'}</p>
  </div>
}

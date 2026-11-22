import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import { trustLevelName, type TrustProgress, type TrustLimits } from '@gooseforum/client'
import { trustQuotaFields } from '@gooseforum/client/trust-limits'
import { useGooseRuntime, useLatestRequest } from '@gooseforum/runtime'
import { useServerErrorMessage } from '@gooseforum/runtime/i18n/server-error'
import { Button } from '@gooseforum/ui/components/button'
import { Spinner } from '@gooseforum/ui/components/spinner'
import { RefreshCw } from 'lucide-react'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@gooseforum/ui/components/table'

export function useTrustProgress() {
  const { api } = useGooseRuntime()
  const { t } = useTranslation('settings')
  const [data, setData] = useState<TrustProgress>()
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [reload, setReload] = useState(0)
  const serverError = useServerErrorMessage()
  const beginRequest = useLatestRequest();
  useEffect(() => {
    const isCurrent = beginRequest()
    if (!api.users.trustLevel) { void Promise.resolve().then(() => { if (isCurrent()) setLoading(false) }); return; }
    void api.users.trustLevel().then(p => { if (isCurrent()) setData(p) }).catch(e => { if (isCurrent()) setError(serverError(e, t('trustProgress.loadFailed'))) }).finally(() => { if (isCurrent()) setLoading(false) })
      }, [api, reload, serverError, beginRequest, t])
  return { data, error, loading, retry: () => { setError(''); setLoading(true); setReload(v => v + 1) } }
}

export function TrustProgressSection({ progress }: { progress: ReturnType<typeof useTrustProgress> }) {
  const { locale } = useGooseRuntime()
  const { t } = useTranslation('settings')
  const { data, error, loading, retry } = progress
  return <section className="space-y-3 border-t p-4">
    <h3 className="text-sm font-semibold">{t('trustProgress.level')}</h3>
    {loading ? <Spinner /> : error ? <div role="alert" className="flex items-center gap-2 text-sm">{error}<Button type="button" size="icon-sm" variant="outline" aria-label={t('trustProgress.retry')} title={t('trustProgress.retry')} onClick={retry}><RefreshCw /></Button></div> : data ? <>
      <div className="text-sm font-medium">{trustLevelName(data.level, locale)}</div>
      {data.locked ? <p className="text-xs text-muted-foreground">{t('trustProgress.locked')}</p> : data.next ? <dl className="grid gap-3 text-sm sm:grid-cols-3">
        <div><dt className="text-xs text-muted-foreground">{t('trustProgress.next')}</dt><dd>{trustLevelName(data.next.level, locale)}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{t('trustProgress.visits')}</dt><dd>{data.visitDays} / {data.next.minVisitDays}</dd></div>
        <div><dt className="text-xs text-muted-foreground">{t('trustProgress.topics')}</dt><dd>{data.topicsVisited} / {data.next.minTopicsVisited}</dd></div>
        <div className="sm:col-span-3"><dt className="text-xs text-muted-foreground">{t('trustProgress.ageMet')}</dt><dd>{new Date(new Date(data.createdAt).getTime() + data.next.minAccountAgeHours * 3600000).toLocaleString(locale)}</dd></div>
      </dl> : null}
      {data.limits ? <TrustCapabilities data={data} limits={data.limits} locale={locale} /> : null}
    </> : null}
  </section>
}

function TrustCapabilities({ data, limits, locale }: { data: TrustProgress; limits: TrustLimits; locale: string }) {
  const [viewedAt] = useState(() => Date.now())
  const { t } = useTranslation('settings')
  const next = !data.locked ? data.nextLimits : null
  const labels = {
    maxDailyTopics: t('trustProgress.dailyTopics'),
    maxDailyReplies: t('trustProgress.dailyReplies'),
    maxDailyUploads: t('trustProgress.dailyUploads'),
    maxImagesPerPost: t('trustProgress.images'),
    maxExternalLinksPerPost: t('trustProgress.links'),
    maxMentionsPerPost: t('trustProgress.mentions'),
  }
  const quota = (value: number | null) => value === null ? (t('trustProgress.unlimited')) : String(value)
  const upload = (value: TrustLimits) => data.communityUploadsEnabled !== false && value.allowImageUpload ? (t('trustProgress.allowed')) : (t('trustProgress.denied'))
  function cooldown(minutes: number) {
    if (minutes === 0) return t('trustProgress.noWait')
    const until = new Date(new Date(data.createdAt).getTime() + minutes * 60000)
    return until.getTime() <= viewedAt ? (t('trustProgress.completed')) : until.toLocaleString(locale)
  }
  return <div className="space-y-3">
    <h4 className="text-sm font-medium">{t('trustProgress.capabilities')}</h4>
    {data.restrictionStatus && data.restrictionStatus !== 'normal' ? <p className="text-sm text-destructive">{t('trustProgress.restricted')}</p> : null}
    <Table className="w-full table-fixed text-xs">
      <TableHeader><TableRow><TableHead className="w-1/2 whitespace-normal">{t('trustProgress.capability')}</TableHead><TableHead className="whitespace-normal">{trustLevelName(data.level, locale)}</TableHead>{next ? <TableHead className="whitespace-normal">{trustLevelName(next.level, locale)}</TableHead> : null}</TableRow></TableHeader>
      <TableBody>
        <TableRow><TableCell className="whitespace-normal">{t('trustProgress.communityUploads')}</TableCell><TableCell>{upload(limits)}</TableCell>{next ? <TableCell>{upload(next) === upload(limits) && next.allowProfileCover === limits.allowProfileCover ? (t('trustProgress.unchanged')) : upload(next)}</TableCell> : null}</TableRow>
        <TableRow><TableCell className="whitespace-normal">{t('trustProgress.cover')}</TableCell><TableCell>{limits.allowProfileCover ? (t('trustProgress.allowed')) : (t('trustProgress.denied'))}</TableCell>{next ? <TableCell>{next.allowProfileCover === limits.allowProfileCover ? (t('trustProgress.unchanged')) : next.allowProfileCover ? (t('trustProgress.allowed')) : (t('trustProgress.denied'))}</TableCell> : null}</TableRow>
        {trustQuotaFields.map(field => <TableRow key={field}><TableCell className="whitespace-normal">{labels[field]}</TableCell><TableCell>{quota(limits[field])}</TableCell>{next ? <TableCell>{next[field] === limits[field] ? (t('trustProgress.unchanged')) : quota(next[field])}</TableCell> : null}</TableRow>)}
      </TableBody>
    </Table>
    {next && trustQuotaFields.every(field => next[field] === limits[field]) && upload(next) === upload(limits) && next.allowProfileCover === limits.allowProfileCover ? <p className="text-xs text-muted-foreground">{t('trustProgress.sameLimits')}</p> : null}
    {data.registrationCooldown ? <dl className="grid gap-3 text-xs sm:grid-cols-2"><div><dt className="text-muted-foreground">{t('trustProgress.postingCooldown')}</dt><dd>{cooldown(data.registrationCooldown.minPostAccountAgeMinutes)}</dd></div><div><dt className="text-muted-foreground">{t('trustProgress.uploadCooldown')}</dt><dd>{cooldown(data.registrationCooldown.minUploadAccountAgeMinutes)}</dd></div></dl> : null}
    {data.communityUploadRoleExempt ? <p className="text-xs text-muted-foreground">{t('trustProgress.exempt')}</p> : null}
    <p className="text-xs text-muted-foreground">{t('trustProgress.limitHint')}</p>
  </div>
}

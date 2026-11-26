import { Button } from "@gooseforum/ui/components/button"
import { RefreshCw } from "lucide-react"
import { useEffect, useState } from 'react'
import type { GooseSiteApi } from '@gooseforum/client'
import { announceUnreadStatus, unreadStatusEvent, announcePublicTopicMax } from '@gooseforum/runtime'

export function SiteStatusNotice({ api, initialDeployment, admin = false }: {
  api: GooseSiteApi
  initialDeployment?: string
  admin?: boolean
}) {
  const [available, setAvailable] = useState('')
  const [dismissed, setDismissed] = useState('')
  const [language, setLanguage] = useState(document.documentElement.lang)
  const zh = language.toLowerCase().startsWith('zh')
  useEffect(() => {
    const update = () => setLanguage(document.documentElement.lang)
    const observer = new MutationObserver(update)
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] })
    update()
    return () => observer.disconnect()
  }, [])
  useEffect(() => {
    // Development HTML has no deployment meta; its first successful response
    // supplies a baseline. Production always compares with the original HTML.
    let baseline = initialDeployment
    let stopped = false
    let unreadGeneration = 0
    const invalidateUnread = () => { unreadGeneration++ }
    let request: AbortController | undefined
    const check = async () => {
      if (stopped || request || document.visibilityState === 'hidden') return
      const generation = unreadGeneration
      const controller = new AbortController()
      request = controller
      const timeout = window.setTimeout(() => controller.abort(), 10_000)
      try {
        const status = await api.status(controller.signal)
        if (stopped || controller.signal.aborted) return
        announcePublicTopicMax(status.publicMaxTopicId)
        if (status.unread && generation === unreadGeneration) announceUnreadStatus(status.unread)
        if (typeof status.deploymentId !== 'string' || !status.deploymentId) return
        baseline ??= status.deploymentId
        if (status.deploymentId !== baseline) setAvailable(status.deploymentId)
      } catch { /* Offline or transient failures must not imply a deployment. */ }
      finally { window.clearTimeout(timeout); request = undefined }
    }
    const visible = () => { if (document.visibilityState === 'visible') void check() }
    const preloadFailed = () => { setAvailable('resource-load-failed'); setDismissed('') }
    window.addEventListener(unreadStatusEvent, invalidateUnread)
    document.addEventListener('visibilitychange', visible)
    window.addEventListener('online', visible)
    window.addEventListener('vite:preloadError', preloadFailed)
    const timer = window.setInterval(() => void check(), 30_000)
    void check()
    return () => {
      stopped = true
      request?.abort()
      window.clearInterval(timer)
      window.removeEventListener(unreadStatusEvent, invalidateUnread)
      document.removeEventListener('visibilitychange', visible)
      window.removeEventListener('online', visible)
      window.removeEventListener('vite:preloadError', preloadFailed)
    }
  }, [api, initialDeployment])
  if (!available || available === dismissed) return null
  const refresh = () => {
    // Main-site editors already guard full reloads via beforeunload. Admin
    // forms do not all have that protection, so require an explicit reminder.
    if (admin && !window.confirm(zh ? '刷新会丢失未保存的修改，请确认已保存。现在刷新？' : 'Refreshing discards unsaved changes. Have you saved your work?')) return
    window.location.reload()
  }
  const failed = available === 'resource-load-failed'
  return <aside role="status" aria-live="polite"
    className="fixed bottom-[max(1rem,env(safe-area-inset-bottom))] left-1/2 z-[10000] w-[calc(100%-2rem)] max-w-md -translate-x-1/2 rounded-xl border bg-card p-4 text-card-foreground shadow-lg">
    <div className="flex items-start gap-3">
      <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary" aria-hidden="true">
        <RefreshCw className="size-4" />
      </span>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-semibold">{zh ? (failed ? '页面需要刷新' : '网站已更新') : (failed ? 'Refresh needed' : 'Update available')}</p>
        <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{zh
          ? (failed ? '页面资源加载失败，请保存未完成的内容后刷新重试。' : '请先保存未完成的内容，再刷新页面使用最新版本。')
          : (failed ? 'Page resources failed to load. Save your work, then refresh to retry.' : 'The site has been updated. Save your work, then refresh.')}</p>
      </div>
    </div>
    <div className="mt-4 flex flex-wrap justify-end gap-2">
      <Button type="button" variant="ghost" size="sm" className="text-muted-foreground" onClick={() => setDismissed(available)}>{zh ? '稍后' : 'Later'}</Button>
      <Button type="button" size="sm" onClick={refresh}>{zh ? '刷新页面' : 'Refresh page'}</Button>
    </div>
  </aside>
}

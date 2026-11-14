import { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
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
    const observer = new MutationObserver(() => setLanguage(document.documentElement.lang))
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['lang'] })
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
  return <aside role="status" aria-live="polite" style={{ position: 'fixed', bottom: 16, left: '50%', transform: 'translateX(-50%)', zIndex: 10000, width: 'min(32rem, calc(100vw - 32px))', boxSizing: 'border-box', padding: 16, borderRadius: 12, background: '#172033', color: '#fff', boxShadow: '0 4px 24px #0004', fontSize: 14 }}>
    <p style={{ margin: '0 0 12px' }}>{zh
      ? (failed ? '页面资源加载失败，请保存未完成的内容后刷新重试。' : '网站已更新，请保存未完成的内容后刷新页面。')
      : (failed ? 'Page resources failed to load. Save your work, then refresh to retry.' : 'The site has been updated. Save your work, then refresh.')}</p>
    <div style={{ display: 'flex', gap: 16 }}>
      <button type="button" onClick={refresh} style={{ cursor: 'pointer', textDecoration: 'underline' }}>{zh ? '刷新页面' : 'Refresh page'}</button>
      <button type="button" onClick={() => setDismissed(available)} style={{ cursor: 'pointer' }}>{zh ? '稍后' : 'Later'}</button>
    </div>
  </aside>
}

// Outside the page root so a failed lazy import or page boundary cannot hide
// the recovery action. Both browser entry points use this one detector.
export function startSiteStatus(api: GooseSiteApi, admin = false) {
  const element = document.createElement('div')
  element.id = 'goose-site-status'
  document.body.append(element)
  const root = createRoot(element)
  const initialDeployment = document.querySelector<HTMLMetaElement>('meta[name="goose-deployment"]')?.content
  root.render(<SiteStatusNotice api={api} initialDeployment={initialDeployment} admin={admin} />)
  return () => { root.unmount(); element.remove() }
}

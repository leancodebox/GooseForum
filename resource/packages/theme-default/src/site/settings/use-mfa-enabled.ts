import { useEffect, useState } from 'react'
import { useGooseRuntime, useLatestRequest } from '@gooseforum/runtime'

export function useMFAEnabled() {
  const runtime = useGooseRuntime()
  const [enabled, setEnabled] = useState<boolean | null>(null)
  const [error, setError] = useState(false)
  const [retry, setRetry] = useState(0)
  const beginRequest = useLatestRequest();
  useEffect(() => {
    const isCurrent = beginRequest()
    void Promise.resolve().then(() => {
      if (!isCurrent()) return
      setEnabled(null); setError(false)
      return runtime.api.users.mfaStatus ? runtime.api.users.mfaStatus() : { enabled: false }
    }).then((status) => {
      if (isCurrent() && status) setEnabled(status.enabled)
    }).catch(() => { if (isCurrent()) setError(true) })
      }, [runtime.api.users, retry, beginRequest])
  return { enabled: enabled === true, ready: enabled !== null, error, retry: () => setRetry((value) => value + 1) }
}

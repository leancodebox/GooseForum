import { useEffect, useState } from 'react'
import { useGooseRuntime } from '@gooseforum/runtime'

export function useMFAEnabled() {
  const runtime = useGooseRuntime()
  const [enabled, setEnabled] = useState<boolean | null>(null)
  const [error, setError] = useState(false)
  const [retry, setRetry] = useState(0)
  useEffect(() => {
    let active = true
    void Promise.resolve().then(() => {
      if (!active) return
      setEnabled(null); setError(false)
      return runtime.api.users.mfaStatus ? runtime.api.users.mfaStatus() : { enabled: false }
    }).then((status) => {
      if (active && status) setEnabled(status.enabled)
    }).catch(() => { if (active) setError(true) })
    return () => { active = false }
  }, [runtime.api.users, retry])
  return { enabled: enabled === true, ready: enabled !== null, error, retry: () => setRetry((value) => value + 1) }
}

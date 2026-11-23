import { useEffect, useRef } from 'react'
import 'altcha'
import type {} from 'altcha/types/react'
import 'altcha/i18n/zh-cn'
import 'altcha/i18n/it'
import 'altcha/i18n/ja'
import 'altcha/i18n/ru'
import type { AuthProof } from './auth-challenge'

export default function AltchaChallenge({ action, proof }: { action: 'login' | 'register'; proof: AuthProof }) {
  const ref = useRef<HTMLElement>(null)
  useEffect(() => {
    const widget = ref.current!
    function onState(event: Event) {
      const { state, payload } = (event as CustomEvent<{ state: string; payload?: string }>).detail
      proof.current = state === 'verified' && payload ? async () => payload : null
    }
    widget.addEventListener('statechange', onState)
    return () => {
      widget.removeEventListener('statechange', onState)
      proof.current = null
    }
  }, [proof])
  return <altcha-widget ref={ref} style={{ '--altcha-max-width': '100%', '--altcha-border-color': 'var(--input)' }} auto="onfocus" challenge={`/api/altcha-challenge?action=${action}`}
    configuration='{"humanInteractionSignature":false}' />
}

import { useEffect, type RefObject } from 'react'
import HCaptcha from '@hcaptcha/react-hcaptcha'
import ReCAPTCHA from 'react-google-recaptcha'
import { GoogleReCaptchaProvider, useGoogleReCaptcha } from 'react-google-recaptcha-v3'

export type AuthProof = RefObject<(() => Promise<string>) | null>

type Action = 'login' | 'register'

function ScoreChallenge({ proof, action }: { proof: AuthProof; action: Action }) {
  const { executeRecaptcha } = useGoogleReCaptcha()
  useEffect(() => {
    proof.current = executeRecaptcha ? () => executeRecaptcha(action) : null
    return () => { proof.current = null }
  }, [executeRecaptcha, proof, action])
  return null
}

interface Props {
  action: Action
  provider: 'hcaptcha' | 'recaptcha_v2' | 'recaptcha_v3'
  siteKey: string
  proof: AuthProof
  onError: () => void
}

export default function AuthChallenge({ provider, siteKey, proof, action, onError }: Props) {
  useEffect(() => () => { proof.current = null }, [proof])
  function setToken(token: string | null) {
    proof.current = token ? async () => token : null
  }
  function fail() {
    setToken(null)
    onError()
  }
  if (provider === 'hcaptcha') {
    return <HCaptcha sitekey={siteKey} reCaptchaCompat={false} onVerify={setToken}
      onExpire={() => setToken(null)} onError={fail} />
  }
  if (provider === 'recaptcha_v2') {
    return <ReCAPTCHA sitekey={siteKey} onChange={setToken} onExpired={() => setToken(null)} onErrored={fail} />
  }
  return <GoogleReCaptchaProvider reCaptchaKey={siteKey}><ScoreChallenge proof={proof} action={action} /></GoogleReCaptchaProvider>
}

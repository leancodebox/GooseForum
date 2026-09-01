import type { GooseSiteApi } from './api/types.js'
import { GooseClientError, GooseProtocolError } from './http/error.js'

export interface PasswordLoginInput {
  username: string
  password: string
  captchaId: string
  captchaCode: string
}

/** Encrypt a password using the server's ephemeral RSA-OAEP public key. */
export async function encryptLoginPassword(
  publicKey: string,
  password: string,
  serverTs: number,
  crypto: Crypto = globalThis.crypto,
) {
  if (!crypto?.subtle) throw new GooseProtocolError('Web Crypto is required for password login')
  const payload = JSON.stringify({ password, ts: serverTs })
  const key = await crypto.subtle.importKey(
    'spki',
    pemToArrayBuffer(publicKey),
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  )
  const encrypted = await crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    key,
    new TextEncoder().encode(payload),
  )
  return arrayBufferToBase64(encrypted)
}

/** Complete the public-key login flow, including one retry for an expired key. */
export async function loginWithPassword(auth: GooseSiteApi['auth'], input: PasswordLoginInput) {
  const submit = async () => {
    const key = await auth.loginPublicKey()
    const encryptedPassword = await encryptLoginPassword(key.publicKey, input.password, key.serverTs)
    await auth.login({
      username: input.username,
      encryptedPassword,
      captchaId: input.captchaId,
      captchaCode: input.captchaCode,
    })
  }
  try {
    await submit()
  } catch (error) {
    if (!(error instanceof GooseClientError) || error.messageCode !== 'auth.login.invalidRequest') throw error
    await submit()
  }
}

function pemToArrayBuffer(pem: string): ArrayBuffer {
  const base64 = pem
    .replace(/-----BEGIN PUBLIC KEY-----/g, '')
    .replace(/-----END PUBLIC KEY-----/g, '')
    .replace(/\s/g, '')
  let binary: string
  try {
    binary = globalThis.atob(base64)
  } catch (cause) {
    throw new GooseProtocolError('GooseForum login public key is invalid', { cause })
  }
  return Uint8Array.from(binary, (char) => char.charCodeAt(0)).buffer
}

function arrayBufferToBase64(buffer: ArrayBuffer) {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return globalThis.btoa(binary)
}

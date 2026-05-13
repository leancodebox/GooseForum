import axiosInstance from './axiosInstance'

let publicKeyPromise: Promise<string> | undefined

type LoginPublicKeyResponse = {
    publicKey: string
    algorithm: string
}

export async function encryptLoginPassword(password: string): Promise<string> {
    if (!window.crypto?.subtle) {
        throw new Error('当前浏览器不支持安全登录加密')
    }

    const publicKey = await getLoginPublicKey()
    const key = await window.crypto.subtle.importKey(
        'spki',
        pemToArrayBuffer(publicKey),
        {
            name: 'RSA-OAEP',
            hash: 'SHA-256',
        },
        false,
        ['encrypt'],
    )

    const payload = JSON.stringify({
        password,
        ts: Date.now(),
    })
    const encrypted = await window.crypto.subtle.encrypt(
        { name: 'RSA-OAEP' },
        key,
        new TextEncoder().encode(payload),
    )

    return arrayBufferToBase64(encrypted)
}

async function getLoginPublicKey(): Promise<string> {
    if (!publicKeyPromise) {
        publicKeyPromise = axiosInstance
            .get('/login-public-key')
            .then((res) => (res.result as LoginPublicKeyResponse).publicKey)
    }
    return publicKeyPromise
}

function pemToArrayBuffer(pem: string): ArrayBuffer {
    const base64 = pem
        .replace(/-----BEGIN PUBLIC KEY-----/g, '')
        .replace(/-----END PUBLIC KEY-----/g, '')
        .replace(/\s/g, '')
    const binary = window.atob(base64)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i += 1) {
        bytes[i] = binary.charCodeAt(i)
    }
    return bytes.buffer
}

function arrayBufferToBase64(buffer: ArrayBuffer): string {
    const bytes = new Uint8Array(buffer)
    let binary = ''
    for (let i = 0; i < bytes.byteLength; i += 1) {
        binary += String.fromCharCode(bytes[i])
    }
    return window.btoa(binary)
}

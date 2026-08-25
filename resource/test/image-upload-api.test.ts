import { File } from 'node:buffer'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { uploadImage } from '../src/runtime/api'

function apiResponse(result: unknown, status = 200) {
  return new Response(JSON.stringify({ code: 0, result }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function apiError(messageCode: string, status: number) {
  return new Response(JSON.stringify({ code: 1, messageCode }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function imageFile() {
  return new File([new Uint8Array([1, 2, 3])], 'image.webp', { type: 'image/webp' }) as unknown as globalThis.File
}

describe('image upload API', () => {
  beforeEach(() => {
    vi.stubGlobal('File', File)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('uses the existing multipart endpoint for the database driver', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({ mode: 'proxy' }))
      .mockResolvedValueOnce(apiResponse({ url: '/file/img/database.webp' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).resolves.toBe('/file/img/database.webp')
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(fetchMock.mock.calls[0][0]).toBe('/file/img-upload/init')
    expect(fetchMock.mock.calls[1][0]).toBe('/file/img-upload')
    expect(fetchMock.mock.calls[1][1]?.body).toBeInstanceOf(FormData)
  })

  it('posts policy fields to object storage and confirms the upload', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct',
        name: '2026/08/25/image.webp',
        upload: {
          url: 'https://objects.example.com/forum',
          method: 'POST',
          fields: { key: '2026/08/25/image.webp', policy: 'signed-policy' },
          expiresAt: '2026-08-25T12:10:00Z',
        },
      }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(apiResponse({ url: '/file/img/2026/08/25/image.webp' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).resolves.toBe('/file/img/2026/08/25/image.webp')
    expect(fetchMock).toHaveBeenCalledTimes(3)
    expect(fetchMock.mock.calls[1][0]).toBe('https://objects.example.com/forum')
    const form = fetchMock.mock.calls[1][1]?.body as FormData
    expect(form.get('key')).toBe('2026/08/25/image.webp')
    expect(form.get('policy')).toBe('signed-policy')
    expect(form.get('file')).toBeInstanceOf(File)
    expect(fetchMock.mock.calls[2][0]).toBe('/file/img-upload/complete')
  })

  it('retries confirmation after a network response loss', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct', name: 'image.webp',
        upload: { url: 'https://objects.example.com/forum', method: 'POST', fields: {}, expiresAt: '' },
      }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockRejectedValueOnce(new TypeError('network response lost'))
      .mockResolvedValueOnce(apiResponse({ url: '/file/img/image.webp' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).resolves.toBe('/file/img/image.webp')
    expect(fetchMock).toHaveBeenCalledTimes(4)
    expect(fetchMock.mock.calls.filter(call => call[0] === '/file/img-upload/complete')).toHaveLength(2)
    expect(fetchMock.mock.calls.some(call => call[0] === '/file/img-upload/abort')).toBe(false)
  })

  it('aborts pending metadata when object storage rejects the upload', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct', name: 'image.webp',
        upload: { url: 'https://objects.example.com/forum', method: 'POST', fields: {}, expiresAt: '' },
      }))
      .mockResolvedValueOnce(new Response(null, { status: 403 }))
      .mockResolvedValueOnce(apiResponse(true))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).rejects.toThrow('HTTP 403')
    expect(fetchMock.mock.calls[2][0]).toBe('/file/img-upload/abort')
  })

  it('confirms before cleanup when the object response is lost', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct', name: 'image.webp',
        upload: { url: 'https://objects.example.com/forum', method: 'POST', fields: {}, expiresAt: '' },
      }))
      .mockRejectedValueOnce(new TypeError('object response lost'))
      .mockResolvedValueOnce(apiResponse({ url: '/file/img/image.webp' }))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).resolves.toBe('/file/img/image.webp')
    expect(fetchMock.mock.calls[2][0]).toBe('/file/img-upload/complete')
    expect(fetchMock.mock.calls.some(call => call[0] === '/file/img-upload/abort')).toBe(false)
  })

  it('leaves ambiguous network failures pending for server cleanup', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct', name: 'image.webp',
        upload: { url: 'https://objects.example.com/forum', method: 'POST', fields: {}, expiresAt: '' },
      }))
      .mockRejectedValueOnce(new TypeError('object response lost'))
      .mockRejectedValueOnce(new TypeError('confirm unavailable'))
      .mockRejectedValueOnce(new TypeError('confirm unavailable'))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).rejects.toThrow('object response lost')
    expect(fetchMock.mock.calls.some(call => call[0] === '/file/img-upload/abort')).toBe(false)
  })

  it('aborts a pending upload when the signing response is incomplete', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({ mode: 'direct', name: 'image.webp' }))
      .mockResolvedValueOnce(apiResponse(true))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).rejects.toThrow()
    expect(fetchMock.mock.calls[1][0]).toBe('/file/img-upload/abort')
  })

  it('keeps pending state when confirmation has a transient server failure', async () => {
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(apiResponse({
        mode: 'direct', name: 'image.webp',
        upload: { url: 'https://objects.example.com/forum', method: 'POST', fields: {}, expiresAt: '' },
      }))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(apiError('upload.saveFailed', 500))
      .mockResolvedValueOnce(apiError('upload.saveFailed', 500))
    vi.stubGlobal('fetch', fetchMock)

    await expect(uploadImage(imageFile())).rejects.toThrow()
    expect(fetchMock.mock.calls.filter(call => call[0] === '/file/img-upload/complete')).toHaveLength(2)
    expect(fetchMock.mock.calls.some(call => call[0] === '/file/img-upload/abort')).toBe(false)
  })
})

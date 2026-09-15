import { GooseClientError } from '../http/error.js'
import type { GooseHttpClient } from '../http/client.js'
import type { ImageUploadInitResult } from './types.js'

export interface ImageUploadResult {
  url: string
  filename?: string
  size?: number
}

export interface ImageUploadRoutes {
  init: string
  complete: string
  abort: string
  proxy: string
}

export async function uploadImage(http: GooseHttpClient, file: File, routes: ImageUploadRoutes): Promise<ImageUploadResult> {
  const init = await http.request<ImageUploadInitResult>(routes.init, {
    method: 'POST',
    json: { filename: file.name, contentType: file.type, size: file.size },
  })
  if (init.mode === 'proxy') return uploadThroughServer(http, file, routes.proxy)
  if (init.mode !== 'direct' || !init.name || !init.upload?.url || init.upload.method !== 'PUT') {
    if (init.name) await abortUpload(http, routes.abort, init.name)
    throw new GooseClientError('GooseForum image upload initialization returned incomplete data')
  }

  let response: Response
  try {
    response = await http.fetch(init.upload.url, {
      method: 'PUT',
      headers: init.upload.headers,
      body: file,
    })
  } catch (cause) {
    try {
      return await completeUpload(http, routes.complete, init.name)
    } catch {
      throw new GooseClientError('Direct image upload could not reach object storage; check network access and bucket CORS', { cause })
    }
  }
  if (!response.ok) {
    await abortUpload(http, routes.abort, init.name)
    throw new GooseClientError(`Image object upload failed with HTTP ${response.status}`, { status: response.status })
  }
  try {
    return await completeUpload(http, routes.complete, init.name)
  } catch (error) {
    const transient = error instanceof TypeError
      || (error instanceof GooseClientError && (error.status || 0) >= 500)
    if (!transient) throw error
    return completeUpload(http, routes.complete, init.name)
  }
}

async function uploadThroughServer(http: GooseHttpClient, file: File, path: string): Promise<ImageUploadResult> {
  const formData = new FormData()
  formData.append('file', file)
  const result = await http.request<ImageUploadResult>(path, { method: 'POST', body: formData })
  if (!result.url) throw new GooseClientError('GooseForum image upload returned no URL')
  return result
}

async function completeUpload(http: GooseHttpClient, path: string, name: string): Promise<ImageUploadResult> {
  const result = await http.request<ImageUploadResult>(path, { method: 'POST', json: { name } })
  if (!result.url) throw new GooseClientError('GooseForum image upload returned no URL')
  return result
}

async function abortUpload(http: GooseHttpClient, path: string, name: string) {
  try {
    await http.request(path, { method: 'POST', json: { name } })
  } catch {
    // Pending uploads also expire server-side; abort remains best effort.
  }
}

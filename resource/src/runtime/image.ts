import { i18n } from './i18n'
import type CompressorType from 'compressorjs'

export interface ImageProcessResult {
  file: File
  converted: boolean
  originalSize: number
  newSize: number
}

export function supportsWebP(): boolean {
  const canvas = document.createElement('canvas')
  canvas.width = 1
  canvas.height = 1
  return canvas.toDataURL('image/webp').startsWith('data:image/webp')
}

export function validateImageFile(file: File, maxSize = 10 * 1024 * 1024): string | null {
  if (!file.type.startsWith('image/')) return i18n.global.t('image.selectFile')
  if (file.size > maxSize) return i18n.global.t('image.maxSize', { size: (maxSize / 1024 / 1024).toFixed(0) })
  return null
}

export async function processImageFile(file: File, quality = 0.85): Promise<ImageProcessResult> {
  const originalSize = file.size
  const targetType = supportsWebP() ? 'image/webp' : file.type || 'image/jpeg'
  const shouldConvert = targetType === 'image/webp' && !file.type.includes('webp')

  try {
    const converted = await compressImage(file, targetType, quality)
    return {
      file: converted,
      converted: shouldConvert || converted.size !== originalSize,
      originalSize,
      newSize: converted.size,
    }
  } catch (error) {
    console.warn(i18n.global.t('image.optimizeFailed'), error)
    return {
      file,
      converted: false,
      originalSize,
      newSize: originalSize,
    }
  }
}

export async function createSquareAvatarFile(file: File, size = 400, quality = 0.86): Promise<File> {
  const bitmap = await createBitmap(file)
  const sourceSize = Math.min(bitmap.width, bitmap.height)
  const sx = Math.floor((bitmap.width - sourceSize) / 2)
  const sy = Math.floor((bitmap.height - sourceSize) / 2)

  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const ctx = canvas.getContext('2d')
  if (!ctx) throw new Error(i18n.global.t('image.processFailed'))

  ctx.imageSmoothingEnabled = true
  ctx.imageSmoothingQuality = 'high'
  ctx.drawImage(bitmap, sx, sy, sourceSize, sourceSize, 0, 0, size, size)

  const blob = await canvasToBlob(canvas, file.type || 'image/png', quality)
  return new File([blob], file.name, {
    type: blob.type || file.type,
    lastModified: Date.now(),
  })
}

export async function canvasToImageFile(
  canvas: HTMLCanvasElement,
  filename: string,
  mimeType = supportsWebP() ? 'image/webp' : 'image/jpeg',
  quality = 0.86,
): Promise<File> {
  const blob = await canvasToBlob(canvas, mimeType, quality)
  return new File([blob], filename.replace(/\.[^/.]+$/, mimeType === 'image/webp' ? '.webp' : '.jpg'), {
    type: mimeType,
    lastModified: Date.now(),
  })
}

async function compressImage(file: File, mimeType: string, quality: number): Promise<File> {
  const { default: Compressor } = await import('compressorjs') as { default: typeof CompressorType }
  return new Promise((resolve, reject) => {
    new Compressor(file, {
      quality,
      mimeType,
      checkOrientation: true,
      convertSize: 0,
      success: (result) => {
        const extension = mimeType === 'image/webp' ? '.webp' : '.jpg'
        if (result instanceof File && result.type === mimeType) {
          resolve(result)
          return
        }
        resolve(new File([result], file.name.replace(/\.[^/.]+$/, extension), {
          type: mimeType,
          lastModified: Date.now(),
        }))
      },
      error: (error) => {
        reject(new Error(i18n.global.t('image.compressFailed', { message: error.message })))
      },
    })
  })
}

async function createBitmap(file: File): Promise<ImageBitmap> {
  if ('createImageBitmap' in window) {
    return createImageBitmap(file, { imageOrientation: 'from-image' })
  }

  const url = URL.createObjectURL(file)
  try {
    const image = await new Promise<HTMLImageElement>((resolve, reject) => {
      const img = new Image()
      img.onload = () => resolve(img)
      img.onerror = () => reject(new Error(i18n.global.t('image.readFailed')))
      img.src = url
    })
    const canvas = document.createElement('canvas')
    canvas.width = image.naturalWidth
    canvas.height = image.naturalHeight
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error(i18n.global.t('image.processFailed'))
    ctx.drawImage(image, 0, 0)
    return createImageBitmap(canvas)
  } finally {
    URL.revokeObjectURL(url)
  }
}

function canvasToBlob(canvas: HTMLCanvasElement, mimeType: string, quality: number): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => {
        if (blob) resolve(blob)
        else reject(new Error(i18n.global.t('image.encodeFailed')))
      },
      mimeType,
      quality,
    )
  })
}

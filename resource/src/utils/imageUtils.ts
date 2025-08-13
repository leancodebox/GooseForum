/**
 * 图片处理工具函数
 * 基于 compressorjs 提供 WebP 压缩功能
 */

import Compressor from 'compressorjs'

/**
 * 图片处理结果接口
 */
export interface ImageProcessResult {
  file: File
  converted: boolean
  originalSize: number
  newSize: number
}

/**
 * 图片信息接口
 */
export interface ImageInfo {
  width: number
  height: number
  size: number
  type: string
  name: string
}


/**
 * 检查浏览器是否支持WebP格式
 * @returns {boolean} 是否支持WebP
 */
export const supportsWebP = (): boolean => {
  const canvas = document.createElement('canvas')
  canvas.width = 1
  canvas.height = 1
  return canvas.toDataURL('image/webp').indexOf('data:image/webp') === 0
}

/**
 * 将图片文件转换为WebP格式
 * @param {File} file - 原始图片文件
 * @param {number} quality - 压缩质量 (0-1之间，默认0.85)
 * @returns {Promise<File>} 转换后的WebP文件
 */
export const convertToWebP = async (file: File, quality: number = 0.85): Promise<File> => {
  return new Promise((resolve, reject) => {
    new Compressor(file, {
      quality,
      mimeType: 'image/webp',
      checkOrientation: true,
      success: (result) => {
        // 确保返回的是 File 对象而不是 Blob
        if (result instanceof File) {
          resolve(result)
        } else {
          // 如果返回的是 Blob，转换为 File
          const webpFile = new File([result], file.name.replace(/\.[^/.]+$/, '.webp'), {
            type: 'image/webp',
            lastModified: Date.now()
          })
          resolve(webpFile)
        }
      },
      error: (error) => {
        reject(new Error(`WebP转换失败: ${error.message}`))
      }
    })
  })
}

/**
 * 智能处理图片文件，如果浏览器支持WebP且文件不是WebP格式则自动转换
 * @param {File} file - 原始图片文件
 * @param {number} quality - WebP压缩质量 (0-1之间，默认0.85)
 * @param {(message: string) => void} onProgress - 进度回调函数
 * @returns {Promise<ImageProcessResult>} 处理结果
 */
export const processImageFile = async (
  file: File, 
  quality: number = 0.85,
  onProgress?: (message: string) => void
): Promise<ImageProcessResult> => {
  const originalSize = file.size
  
  // 检查是否需要转换为WebP
  const shouldConvert = supportsWebP() && !file.type.includes('webp')
  
  if (shouldConvert) {
    try {
      onProgress?.('正在优化图片格式...')
      const convertedFile = await convertToWebP(file, quality)
      
      const compressionRatio = ((originalSize - convertedFile.size) / originalSize * 100).toFixed(1)
       console.log(`图片已转换为WebP格式，原大小: ${(originalSize / 1024).toFixed(1)}KB，转换后: ${(convertedFile.size / 1024).toFixed(1)}KB，压缩率: ${compressionRatio}%`)
      
      return {
        file: convertedFile,
        converted: true,
        originalSize,
        newSize: convertedFile.size
      }
    } catch (error) {
      console.warn('WebP转换失败，使用原始文件:', error)
    }
  }
  
  // 不需要转换、不支持WebP或转换失败时返回原始文件
  return {
    file,
    converted: false,
    originalSize,
    newSize: originalSize
  }
}

/**
 * 图片文件验证
 * @param {File} file - 要验证的文件
 * @param {number} maxSize - 最大文件大小（字节）
 * @returns {string | null} 错误信息，null表示验证通过
 */
export const validateImageFile = (file: File, maxSize: number = 10 * 1024 * 1024): string | null => {
  if (!file.type.startsWith('image/')) {
    return '请选择图片文件'
  }
  
  if (file.size > maxSize) {
    return `图片大小不能超过${(maxSize / 1024 / 1024).toFixed(0)}MB`
  }
  
  return null
}
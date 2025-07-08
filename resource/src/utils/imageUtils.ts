/**
 * 图片处理工具函数
 * 提供WebP转换、浏览器兼容性检测等功能
 */

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
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      canvas.width = img.width
      canvas.height = img.height
      
      if (ctx) {
        ctx.drawImage(img, 0, 0)
        
        canvas.toBlob((blob) => {
          if (blob) {
            const webpFile = new File([blob], file.name.replace(/\.[^/.]+$/, '.webp'), {
              type: 'image/webp',
              lastModified: Date.now()
            })
            resolve(webpFile)
          } else {
            reject(new Error('WebP转换失败'))
          }
        }, 'image/webp', quality)
      } else {
        reject(new Error('Canvas上下文获取失败'))
      }
    }
    
    img.onerror = () => {
      reject(new Error('图片加载失败'))
    }
    
    // 创建对象URL并设置给图片
    const objectUrl = URL.createObjectURL(file)
    img.src = objectUrl
    
    // 清理对象URL以避免内存泄漏
    img.onload = () => {
      URL.revokeObjectURL(objectUrl)
      canvas.width = img.width
      canvas.height = img.height
      
      if (ctx) {
        ctx.drawImage(img, 0, 0)
        
        canvas.toBlob((blob) => {
          if (blob) {
            const webpFile = new File([blob], file.name.replace(/\.[^/.]+$/, '.webp'), {
              type: 'image/webp',
              lastModified: Date.now()
            })
            resolve(webpFile)
          } else {
            reject(new Error('WebP转换失败'))
          }
        }, 'image/webp', quality)
      } else {
        reject(new Error('Canvas上下文获取失败'))
      }
    }
  })
}

/**
 * 智能处理图片文件，如果浏览器支持WebP且文件不是WebP格式则自动转换
 * @param {File} file - 原始图片文件
 * @param {number} quality - WebP压缩质量 (0-1之间，默认0.85)
 * @param {(message: string) => void} onProgress - 进度回调函数
 * @returns {Promise<{file: File, converted: boolean, originalSize: number, newSize: number}>}
 */
export const processImageFile = async (
  file: File, 
  quality: number = 0.85,
  onProgress?: (message: string) => void
): Promise<{
  file: File
  converted: boolean
  originalSize: number
  newSize: number
}> => {
  const originalSize = file.size
  
  // 如果浏览器支持WebP且文件不是WebP格式，则转换为WebP
  if (supportsWebP() && !file.type.includes('webp')) {
    try {
      onProgress?.('正在优化图片格式...')
      const convertedFile = await convertToWebP(file, quality)
      
      console.log(`图片已转换为WebP格式，原大小: ${(originalSize / 1024).toFixed(1)}KB，转换后: ${(convertedFile.size / 1024).toFixed(1)}KB`)
      
      return {
        file: convertedFile,
        converted: true,
        originalSize,
        newSize: convertedFile.size
      }
    } catch (error) {
      console.warn('WebP转换失败，使用原始文件:', error)
      // 转换失败时使用原始文件
      return {
        file,
        converted: false,
        originalSize,
        newSize: originalSize
      }
    }
  }
  
  // 不需要转换或不支持WebP
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
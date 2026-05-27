import { nextTick, onBeforeUnmount, ref } from 'vue'
import type Cropper from 'cropperjs'
import { uploadAvatar } from '@/runtime/api'
import { canvasToImageFile, validateImageFile } from '@/runtime/image'

interface AvatarCropUploadOptions {
  initialAvatarUrl: string
  onStatus: (message: string) => void
  onError: (message: string) => void
}

export function useAvatarCropUpload(options: AvatarCropUploadOptions) {
  const uploadingAvatar = ref(false)
  const avatarInput = ref<HTMLInputElement | null>(null)
  const cropperImage = ref<HTMLImageElement | null>(null)
  const avatarUrl = ref(options.initialAvatarUrl)
  const cropModalOpen = ref(false)
  const cropImageUrl = ref('')
  const cropPreviewUrl = ref('')
  const cropError = ref('')
  const cropSourceFile = ref<File | null>(null)
  let cropper: Cropper | undefined
  let cropperContainer: HTMLElement | undefined
  let cropPreviewFrame = 0

  onBeforeUnmount(() => {
    destroyCropper()
    revokeCropImageUrl()
  })

  function chooseAvatar() {
    avatarInput.value?.click()
  }

  async function handleAvatarChange(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0]
    if (!file) return
    const validationError = validateImageFile(file, 5 * 1024 * 1024)
    if (validationError) return options.onError(validationError)

    openCropModal(file)
    if (avatarInput.value) avatarInput.value.value = ''
  }

  function openCropModal(file: File) {
    destroyCropper()
    revokeCropImageUrl()
    cropError.value = ''
    cropSourceFile.value = file
    cropImageUrl.value = URL.createObjectURL(file)
    cropModalOpen.value = true
    void nextTick(() => {
      void initCropper()
    })
  }

  async function initCropper() {
    const image = cropperImage.value
    if (!image) return
    const { default: Cropper } = await import('cropperjs')
    if (!cropModalOpen.value || cropperImage.value !== image) return

    cropper = new Cropper(image, {
      template: `
        <cropper-canvas background>
          <cropper-image translatable scalable rotatable></cropper-image>
          <cropper-shade hidden></cropper-shade>
          <cropper-handle action="select" plain></cropper-handle>
          <cropper-selection aspect-ratio="1" movable resizable zoomable outlined>
            <cropper-grid role="grid" bordered covered></cropper-grid>
            <cropper-crosshair centered></cropper-crosshair>
            <cropper-handle action="move" theme-color="rgba(37, 99, 235, 0.35)"></cropper-handle>
            <cropper-handle action="n-resize"></cropper-handle>
            <cropper-handle action="e-resize"></cropper-handle>
            <cropper-handle action="s-resize"></cropper-handle>
            <cropper-handle action="w-resize"></cropper-handle>
            <cropper-handle action="ne-resize"></cropper-handle>
            <cropper-handle action="nw-resize"></cropper-handle>
            <cropper-handle action="se-resize"></cropper-handle>
            <cropper-handle action="sw-resize"></cropper-handle>
          </cropper-selection>
        </cropper-canvas>
      `,
    })
    void resetCropSelectionToImageShortSide()
    const container = cropper.container as HTMLElement
    cropperContainer = container
    container.addEventListener('pointerup', scheduleCropPreview)
    container.addEventListener('wheel', scheduleCropPreview, { passive: true })
    container.addEventListener('keyup', scheduleCropPreview)
  }

  async function resetCropSelectionToImageShortSide() {
    const cropperImageElement = cropper?.getCropperImage()
    const cropperCanvas = cropper?.getCropperCanvas()
    const selection = cropper?.getCropperSelection()
    if (!cropperImageElement || !cropperCanvas || !selection) return

    try {
      await cropperImageElement.$ready()
    } catch {
      return
    }

    window.requestAnimationFrame(() => {
      const canvasRect = cropperCanvas.getBoundingClientRect()
      const imageRect = cropperImageElement.getBoundingClientRect()
      const side = Math.min(imageRect.width, imageRect.height)
      if (side <= 0) return

      const x = imageRect.left - canvasRect.left + (imageRect.width - side) / 2
      const y = imageRect.top - canvasRect.top + (imageRect.height - side) / 2
      selection.$change(x, y, side, side, 1, true)
      void updateCropPreview()
    })
  }

  function closeCropModal() {
    cropModalOpen.value = false
    cropSourceFile.value = null
    cropError.value = ''
    destroyCropper()
    revokeCropImageUrl()
  }

  function destroyCropper() {
    window.cancelAnimationFrame(cropPreviewFrame)
    cropPreviewFrame = 0
    if (cropperContainer) {
      cropperContainer.removeEventListener('pointerup', scheduleCropPreview)
      cropperContainer.removeEventListener('wheel', scheduleCropPreview)
      cropperContainer.removeEventListener('keyup', scheduleCropPreview)
      cropperContainer = undefined
    }
    cropper?.destroy()
    cropper = undefined
    cropPreviewUrl.value = ''
  }

  function revokeCropImageUrl() {
    if (cropImageUrl.value) URL.revokeObjectURL(cropImageUrl.value)
    cropImageUrl.value = ''
  }

  async function uploadCroppedAvatar() {
    if (!cropper || !cropSourceFile.value) return

    uploadingAvatar.value = true
    cropError.value = ''
    try {
      const selection = cropper.getCropperSelection()
      if (!selection) throw new Error('请选择裁切区域')
      const canvas = await selection.$toCanvas({
        width: 300,
        height: 300,
        beforeDraw(context) {
          context.imageSmoothingEnabled = true
          context.imageSmoothingQuality = 'high'
        },
      })
      const avatarFiles = await createAvatarUploadFiles(canvas, cropSourceFile.value.name)
      avatarUrl.value = await uploadAvatar(avatarFiles)
      closeCropModal()
      options.onStatus('头像已更新')
    } catch (err) {
      const message = err instanceof Error ? err.message : '头像上传失败'
      cropError.value = message
      options.onError(message)
    } finally {
      uploadingAvatar.value = false
    }
  }

  function scheduleCropPreview() {
    window.cancelAnimationFrame(cropPreviewFrame)
    cropPreviewFrame = window.requestAnimationFrame(() => {
      void updateCropPreview()
    })
  }

  async function updateCropPreview() {
    const selection = cropper?.getCropperSelection()
    if (!selection) return
    try {
      const canvas = await selection.$toCanvas({
        width: 160,
        height: 160,
        beforeDraw(context) {
          context.imageSmoothingEnabled = true
          context.imageSmoothingQuality = 'high'
        },
      })
      cropPreviewUrl.value = canvas.toDataURL('image/webp', 0.82)
    } catch {
      cropPreviewUrl.value = ''
    }
  }

  return {
    uploadingAvatar,
    avatarInput,
    cropperImage,
    avatarUrl,
    cropModalOpen,
    cropImageUrl,
    cropPreviewUrl,
    cropError,
    chooseAvatar,
    handleAvatarChange,
    closeCropModal,
    uploadCroppedAvatar,
  }
}

async function createAvatarUploadFiles(sourceCanvas: HTMLCanvasElement, filename: string): Promise<File[]> {
  const avatar300 = await canvasToImageFile(sourceCanvas, filename, undefined, 0.86)
  const avatarMedium = await canvasToImageFile(resizeAvatarCanvas(sourceCanvas, 96), 'avatar_medium.webp', undefined, 0.9)
  return [avatar300, avatarMedium]
}

function resizeAvatarCanvas(sourceCanvas: HTMLCanvasElement, size: number): HTMLCanvasElement {
  const canvas = document.createElement('canvas')
  canvas.width = size
  canvas.height = size
  const context = canvas.getContext('2d')
  if (!context) throw new Error('无法处理头像')

  context.imageSmoothingEnabled = true
  context.imageSmoothingQuality = 'high'
  context.drawImage(sourceCanvas, 0, 0, sourceCanvas.width, sourceCanvas.height, 0, 0, size, size)
  return canvas
}

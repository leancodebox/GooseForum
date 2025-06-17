<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { VueCropper } from 'vue-cropper'
import 'vue-cropper/dist/index.css'
import { uploadAvatar } from '../utils/articleService.ts'

const props = defineProps({
  currentAvatar: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['avatar-updated'])

// 状态变量
const uploading = ref(false)
const showCropModal = ref(false)
const cropperRef = ref(null)
const previewUrl = ref('')
const cropImg = ref('')
const fileInputRef = ref(null)
const isSmallScreen = ref(false)
const previewLarge = ref(null)
const previewSmall = ref(null)

// 触发文件选择
const triggerFileSelect = () => {
  if (fileInputRef.value && !uploading.value) {
    fileInputRef.value.click()
  }
}

// 处理文件选择
const handleFileSelect = (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 验证文件类型
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif']
  if (!allowedTypes.includes(file.type)) {
    alert('只支持 jpg、png、gif 格式的图片')
    // 重置文件输入框
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
    return
  }

  // 验证文件大小（2MB）
  if (file.size > 2 * 1024 * 1024) {
    alert('图片大小不能超过 2MB')
    // 重置文件输入框
    if (fileInputRef.value) {
      fileInputRef.value.value = ''
    }
    return
  }

  // 清理之前的URL
  if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }

  // 使用FileReader读取文件
  const reader = new FileReader()
  reader.onload = (e) => {
    const result = e.target?.result
    if (result && typeof result === 'string') {
      previewUrl.value = result
      cropImg.value = result
      showCropModal.value = true
    }
  }
  reader.readAsDataURL(file)
}

// 图片压缩函数
const compressImage = (base64Data, maxWidth = 200) => {
  return new Promise((resolve) => {
    const img = new Image()
    img.src = base64Data
    img.onload = () => {
      const canvas = document.createElement('canvas')
      const ctx = canvas.getContext('2d')
      if (ctx == null) {
        resolve(new Blob())
        return
      }

      // 计算压缩后的尺寸，保持宽高比
      let width = img.width
      let height = img.height
      if (width > maxWidth) {
        height = (maxWidth * height) / width
        width = maxWidth
      }

      canvas.width = width
      canvas.height = height

      // 绘制图片
      ctx.drawImage(img, 0, 0, width, height)

      // 转换为 blob，使用较低的质量值来减小文件大小
      canvas.toBlob(
          (blob) => resolve(blob || new Blob()),
          'image/jpeg',
          0.8  // 压缩质量，0.8通常是质量和大小的好平衡点
      )
    }
  })
}

// 裁切完成函数
const handleCropFinish = async () => {
  if (!cropperRef.value) return

  try {
    cropperRef.value.getCropData(async (base64Data) => {
      uploading.value = true
      try {
        // 压缩裁切后的图片
        const compressedBlob = await compressImage(base64Data)

        // 检查压缩后的大小
        if (compressedBlob.size > 500 * 1024) { // 500KB 限制
          alert('图片太大，请选择更小的区域或更小的图片')
          uploading.value = false
          return
        }

        // 创建FormData
        const formData = new FormData()
        formData.append('avatar', compressedBlob, 'avatar.jpg')

        const response = await uploadAvatar(formData)
        if (response.code === 0) {
          emit('avatar-updated', response.result.avatarUrl || response.data?.avatarUrl)
          alert('头像上传成功')
          
          // 清理blob URL资源（如果有的话）
          if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
            URL.revokeObjectURL(previewUrl.value)
          }
          
          // 重置状态
          showCropModal.value = false
          previewUrl.value = ''
          cropImg.value = ''
          
          // 重置文件输入框
          if (fileInputRef.value) {
            fileInputRef.value.value = ''
          }
        }
      } catch (error) {
        alert('头像上传失败')
        console.error('上传失败:', error)
      } finally {
        uploading.value = false
      }
    })
  } catch (err) {
    console.error('裁切失败:', err)
    alert('裁切失败')
  }
}

// 实时预览函数
const realTimePreview = (data) => {
  // VueCropper的realTime事件传递的是裁切后的canvas数据
  if (cropperRef.value) {
    cropperRef.value.getCropData((cropData) => {
      if (cropData) {
        // 更新大预览
        if (previewLarge.value) {
          previewLarge.value.style.backgroundImage = `url(${cropData})`
        }
        // 更新小预览
        if (previewSmall.value) {
          previewSmall.value.style.backgroundImage = `url(${cropData})`
        }
      }
    })
  }
}

// 检查屏幕尺寸
const checkScreenSize = () => {
  isSmallScreen.value = window.innerWidth < 800
}

// 关闭模态框
const closeCropModal = () => {
  showCropModal.value = false
  
  // 清理blob URL资源（如果有的话）
  if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }
  
  // 重置状态
  previewUrl.value = ''
  cropImg.value = ''
  
  // 重置文件输入框
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

// 组件挂载时初始化
onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

// 组件卸载时清理资源
onUnmounted(() => {
  // 只清理blob URL，不清理base64数据
  if (previewUrl.value && previewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewUrl.value)
  }
  window.removeEventListener('resize', checkScreenSize)
})
</script>

<template>
  <div>
    <!-- 头像上传区域 -->
    <div class="grid grid-cols-1 gap-2">
      <!-- 隐藏的文件输入框 -->
      <input 
        ref="fileInputRef" 
        type="file" 
        class="hidden" 
        accept="image/*"
        @change="handleFileSelect" 
        :disabled="uploading"
      />
      
      <!-- 可点击的头像区域 -->
       <div class="avatar">
         <div 
           @click="triggerFileSelect"
           class="mask mask-squircle w-20 h-20 relative cursor-pointer hover:opacity-80 transition-opacity duration-200 group"
           :class="{ 'cursor-not-allowed opacity-50': uploading }"
         >
           <img 
             :src="currentAvatar || '/default-avatar.png'" 
             alt="点击更换头像"
             class="w-full h-full object-cover"
           />
           <!-- 悬停时显示的遮罩层 -->
           <div class="absolute inset-0 bg-black bg-opacity-50 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200">
             <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path>
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path>
             </svg>
           </div>
         </div>
       </div>
      
      <div class="text-sm text-base-content/60">
        点击头像更换 • 支持 JPG、PNG、GIF 格式，文件大小不超过 2MB
      </div>
      <div v-if="uploading" class="text-sm text-primary">
        正在上传头像...
      </div>
    </div>

    <!-- 裁切模态框 - 使用 daisyui Modal -->
    <div class="modal" :class="{ 'modal-open': showCropModal }">
      <div class="modal-box max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <h3 class="text-lg font-normal mb-4">裁切头像</h3>
        
        <div class="flex gap-6">
          <!-- 裁切区域 -->
          <div class="flex-1">
            <h4 class="text-sm font-normal mb-2">调整裁切区域</h4>
            <VueCropper
               ref="cropperRef"
               :img="cropImg"
               :outputSize="1"
               :outputType="'jpeg'"
               :info="true"
               :full="false"
               :canMove="true"
               :canMoveBox="true"
               :original="false"
               :autoCrop="true"
               :fixed="true"
               :fixedNumber="[1, 1]"
               :centerBox="true"
               :infoTrue="true"
               :fixedBox="false"
               :enlarge="1"
               :mode="'contain'"
               :high="true"
               :canScale="true"
               @realTime="realTimePreview"
               style="width: 100%; height: 300px; background: #f5f5f5;"
             />
          </div>
          
          <!-- 预览区域 -->
          <div class="w-1/3">
            <h4 class="text-sm font-normal mb-2">预览效果</h4>
            <div class="space-y-4">
              <div class="text-center">
                <p class="text-xs text-gray-500 mb-2">大头像 (100x100)</p>
                <div ref="previewLarge" class="w-24 h-24 mx-auto rounded-full border-2 border-gray-200 bg-gray-100 bg-cover bg-center"></div>
              </div>
              <div class="text-center">
                <p class="text-xs text-gray-500 mb-2">小头像 (50x50)</p>
                <div ref="previewSmall" class="w-12 h-12 mx-auto rounded-full border-2 border-gray-200 bg-gray-100 bg-cover bg-center"></div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 操作按钮 -->
        <div class="modal-action">
          <button 
            type="button"
            @click="closeCropModal" 
            class="btn btn-outline"
          >
            取消
          </button>
          <button 
            type="button"
            @click="handleCropFinish" 
            class="btn btn-primary"
            :disabled="uploading"
          >
            <span v-if="uploading" class="loading loading-spinner loading-sm"></span>
            {{ uploading ? '上传中...' : '裁切并上传' }}
          </button>
        </div>
      </div>
      <div class="modal-backdrop" @click="closeCropModal"></div>
    </div>
  </div>
</template>
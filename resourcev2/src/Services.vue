<template>
  <div class="services-interactive">
    <div class="max-w-4xl mx-auto px-4 py-8">
      <h3 class="text-2xl font-bold text-gray-900 mb-6 text-center">服务咨询</h3>
      
      <div class="bg-white rounded-lg shadow-lg p-6">
        <form @submit.prevent="submitInquiry" class="space-y-4">
          <div class="grid md:grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">姓名</label>
              <input 
                v-model="inquiry.name" 
                type="text" 
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入您的姓名"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
              <input 
                v-model="inquiry.email" 
                type="email" 
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="请输入您的邮箱"
              >
            </div>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">感兴趣的服务</label>
            <select 
              v-model="inquiry.service" 
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">请选择服务类型</option>
              <option value="web-development">Web开发</option>
              <option value="mobile-app">移动应用开发</option>
              <option value="consulting">技术咨询</option>
              <option value="custom">定制开发</option>
            </select>
          </div>
          
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">详细需求</label>
            <textarea 
              v-model="inquiry.message" 
              rows="4" 
              required
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="请详细描述您的需求..."
            ></textarea>
          </div>
          
          <div class="text-center">
            <button 
              type="submit" 
              :disabled="isSubmitting"
              class="bg-gradient-to-r from-green-500 to-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:from-green-600 hover:to-blue-700 transition-all duration-200 disabled:opacity-50"
            >
              {{ isSubmitting ? '提交中...' : '提交咨询' }}
            </button>
          </div>
        </form>
        
        <div v-if="submitMessage" class="mt-4 p-4 rounded-md" :class="submitSuccess ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'">
          {{ submitMessage }}
        </div>
      </div>
      
      <div class="mt-8 grid md:grid-cols-3 gap-6">
        <div class="text-center p-4 bg-white rounded-lg shadow">
          <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-3">
            <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <h4 class="font-semibold text-gray-900">快速响应</h4>
          <p class="text-sm text-gray-600 mt-1">24小时内回复</p>
        </div>
        
        <div class="text-center p-4 bg-white rounded-lg shadow">
          <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
            <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
          </div>
          <h4 class="font-semibold text-gray-900">专业团队</h4>
          <p class="text-sm text-gray-600 mt-1">经验丰富的开发者</p>
        </div>
        
        <div class="text-center p-4 bg-white rounded-lg shadow">
          <div class="w-12 h-12 bg-purple-100 rounded-full flex items-center justify-center mx-auto mb-3">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
          </div>
          <h4 class="font-semibold text-gray-900">高效交付</h4>
          <p class="text-sm text-gray-600 mt-1">按时完成项目</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import axios from 'axios'

export default {
  name: 'Services',
  setup() {
    const inquiry = ref({
      name: '',
      email: '',
      service: '',
      message: ''
    })
    
    const isSubmitting = ref(false)
    const submitMessage = ref('')
    const submitSuccess = ref(false)
    
    const submitInquiry = async () => {
      isSubmitting.value = true
      submitMessage.value = ''
      
      try {
        const response = await axios.post('/api/service-inquiry', inquiry.value)
        
        if (response.status === 200) {
          submitMessage.value = '咨询提交成功！我们会尽快与您联系。'
          submitSuccess.value = true
          
          // 重置表单
          inquiry.value = {
            name: '',
            email: '',
            service: '',
            message: ''
          }
        }
      } catch (error) {
        submitMessage.value = '提交失败，请稍后重试。'
        submitSuccess.value = false
        console.error('提交咨询失败:', error)
      } finally {
        isSubmitting.value = false
        
        // 3秒后清除消息
        setTimeout(() => {
          submitMessage.value = ''
        }, 3000)
      }
    }
    
    return {
      inquiry,
      isSubmitting,
      submitMessage,
      submitSuccess,
      submitInquiry
    }
  }
}
</script>

<style scoped>
.services-interactive {
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  min-height: 400px;
}
</style>
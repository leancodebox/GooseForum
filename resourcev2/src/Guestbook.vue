<template>
  <div class="min-h-screen bg-gray-50">
    <!-- 页面标题 -->
    <div class="bg-gradient-to-r from-green-600 to-blue-600 text-white py-16">
      <div class="max-w-4xl mx-auto px-4 text-center">
        <h1 class="text-4xl font-bold mb-4">留言板</h1>
        <p class="text-xl">分享您的想法和建议</p>
      </div>
    </div>

    <!-- 留言表单 -->
    <div class="max-w-4xl mx-auto px-4 py-8">
      <div class="bg-white rounded-lg shadow-md p-6 mb-8">
        <h2 class="text-2xl font-bold text-gray-900 mb-6">发表留言</h2>
        <form @submit.prevent="submitMessage" class="space-y-4">
          <div class="grid md:grid-cols-2 gap-4">
            <div>
              <label for="name" class="block text-sm font-medium text-gray-700 mb-2">姓名</label>
              <input 
                type="text" 
                id="name" 
                v-model="newMessage.name" 
                required
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="请输入您的姓名"
              >
            </div>
            <div>
              <label for="email" class="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
              <input 
                type="email" 
                id="email" 
                v-model="newMessage.email" 
                class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                placeholder="请输入您的邮箱（可选）"
              >
            </div>
          </div>
          <div>
            <label for="message" class="block text-sm font-medium text-gray-700 mb-2">留言内容</label>
            <textarea 
              id="message" 
              v-model="newMessage.content" 
              required
              rows="4"
              class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="请输入您的留言内容..."
            ></textarea>
          </div>
          <div class="flex justify-end">
            <button 
              type="submit" 
              :disabled="isSubmitting"
              class="bg-blue-600 text-white px-6 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <span v-if="isSubmitting">提交中...</span>
              <span v-else>发表留言</span>
            </button>
          </div>
        </form>
      </div>

      <!-- 留言列表 -->
      <div class="bg-white rounded-lg shadow-md p-6">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-2xl font-bold text-gray-900">留言列表</h2>
          <div class="text-sm text-gray-500">
            共 {{ messages.length }} 条留言
          </div>
        </div>
        
        <!-- 加载状态 -->
        <div v-if="isLoading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="mt-2 text-gray-600">加载中...</p>
        </div>
        
        <!-- 留言项 -->
        <div v-else-if="messages.length > 0" class="space-y-4">
          <div 
            v-for="(message, index) in messages" 
            :key="message.id || index"
            class="border-b border-gray-200 pb-4 last:border-b-0"
          >
            <div class="flex justify-between items-start mb-2">
              <div class="flex items-center space-x-2">
                <div class="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                  <span class="text-blue-600 font-semibold text-sm">{{ message.name.charAt(0).toUpperCase() }}</span>
                </div>
                <div>
                  <h4 class="font-semibold text-gray-900">{{ message.name }}</h4>
                  <p class="text-sm text-gray-500">{{ formatDate(message.created_at) }}</p>
                </div>
              </div>
              <div v-if="message.email" class="text-sm text-gray-400">
                {{ message.email }}
              </div>
            </div>
            <p class="text-gray-700 leading-relaxed ml-10">{{ message.content }}</p>
          </div>
        </div>
        
        <!-- 空状态 -->
        <div v-else class="text-center py-12">
          <div class="text-gray-400 mb-4">
            <svg class="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"></path>
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-900 mb-2">暂无留言</h3>
          <p class="text-gray-500">成为第一个留言的人吧！</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, reactive, onMounted } from 'vue'
import axios from 'axios'

export default {
  name: 'Guestbook',
  setup() {
    const messages = ref([])
    const isLoading = ref(true)
    const isSubmitting = ref(false)
    
    const newMessage = reactive({
      name: '',
      email: '',
      content: ''
    })

    // 加载留言列表
    const loadMessages = async () => {
      try {
        isLoading.value = true
        const response = await axios.get('/api/messages')
        messages.value = response.data || []
      } catch (error) {
        console.error('加载留言失败:', error)
        // 如果API调用失败，使用空数组
        messages.value = []
      } finally {
        isLoading.value = false
      }
    }

    // 提交留言
    const submitMessage = async () => {
      if (!newMessage.name.trim() || !newMessage.content.trim()) {
        alert('请填写姓名和留言内容')
        return
      }

      try {
        isSubmitting.value = true
        const response = await axios.post('/api/messages', {
          name: newMessage.name.trim(),
          email: newMessage.email.trim(),
          content: newMessage.content.trim()
        })
        
        if (response.data && response.data.data) {
          // 将新留言添加到列表顶部
          messages.value.unshift(response.data.data)
          
          // 清空表单
          newMessage.name = ''
          newMessage.email = ''
          newMessage.content = ''
          
          alert('留言发表成功！')
        }
      } catch (error) {
        console.error('提交留言失败:', error)
        alert('提交失败，请稍后重试')
      } finally {
        isSubmitting.value = false
      }
    }

    // 格式化日期
    const formatDate = (dateStr) => {
      if (!dateStr) return '未知时间'
      if (dateStr === '刚刚') return dateStr
      
      try {
        const date = new Date(dateStr)
        return date.toLocaleString('zh-CN')
      } catch {
        return dateStr
      }
    }

    // 组件挂载时加载数据
    onMounted(() => {
      loadMessages()
    })

    return {
      messages,
      isLoading,
      isSubmitting,
      newMessage,
      submitMessage,
      formatDate
    }
  }
}
</script>

<style scoped>
/* 组件特定样式 */
.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
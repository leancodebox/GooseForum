<template>
  <div class="max-w-4xl mx-auto">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex items-center gap-3 mb-6">
          <a href="/links" class="btn btn-ghost btn-sm">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            返回
          </a>
          <h1 class="card-title text-3xl font-normal text-base-content">申请友情链接</h1>
        </div>
        
        <div class="alert alert-info alert-dash mb-6">
          <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <div>
            <h3 class="font-normal">申请须知</h3>
            <div class="text-sm mt-1">
              请确保您的网站内容健康、有价值，且能够稳定访问。我们会在3-7个工作日内审核您的申请。
            </div>
          </div>
        </div>

        <form @submit.prevent="submitApplication" class="grid grid-cols-1 gap-8">
          <!-- 网站基本信息 -->
          <div class="grid grid-cols-1 gap-6">
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">网站名称 <span class="text-error">*</span></span>
                </label>
                <input 
                  v-model="form.siteName" 
                  type="text" 
                  placeholder="请输入网站名称" 
                  class="input input-bordered w-full"
                  :class="{ 'input-error': errors.siteName }"
                  required
                />
                <label v-if="errors.siteName" class="label">
                  <span class="label-text-alt text-error">{{ errors.siteName }}</span>
                </label>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">网站地址 <span class="text-error">*</span></span>
                </label>
                <input 
                  v-model="form.siteUrl" 
                  type="url" 
                  placeholder="https://example.com" 
                  class="input input-bordered w-full"
                  :class="{ 'input-error': errors.siteUrl }"
                  required
                />
                <label v-if="errors.siteUrl" class="label">
                  <span class="label-text-alt text-error">{{ errors.siteUrl }}</span>
                </label>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">网站Logo</span>
                </label>
                <input 
                  v-model="form.siteLogo" 
                  type="url" 
                  placeholder="https://example.com/logo.png" 
                  class="input input-bordered w-full"
                  :class="{ 'input-error': errors.siteLogo }"
                />
                <label class="label">
                  <span class="label-text-alt">建议尺寸：64x64px，支持 PNG、JPG、SVG 格式</span>
                </label>
                <label v-if="errors.siteLogo" class="label">
                  <span class="label-text-alt text-error">{{ errors.siteLogo }}</span>
                </label>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">网站分类 <span class="text-error">*</span></span>
                </label>
                <select 
                  v-model="form.category" 
                  class="select select-bordered w-full"
                  :class="{ 'select-error': errors.category }"
                  required
                >
                  <option value="">请选择分类</option>
                  <option value="tech-community">技术社区</option>
                  <option value="dev-tools">开发工具</option>
                  <option value="personal-blog">个人博客</option>
                  <option value="other">其他</option>
                </select>
                <label v-if="errors.category" class="label">
                  <span class="label-text-alt text-error">{{ errors.category }}</span>
                </label>
              </div>
            </div>

            <div class="form-control lg:col-span-2">
              <label class="label">
                <span class="label-text font-normal">网站描述 <span class="text-error">*</span></span>
              </label>
              <textarea 
                v-model="form.description" 
                class="textarea textarea-bordered h-24 w-full"
                :class="{ 'textarea-error': errors.description }"
                placeholder="请简要描述您的网站内容和特色（50-200字）"
                required
              ></textarea>
              <label class="label">
                <span class="label-text-alt">{{ form.description.length }}/200</span>
              </label>
              <label v-if="errors.description" class="label">
                <span class="label-text-alt text-error">{{ errors.description }}</span>
              </label>
            </div>

            <!-- 互链信息（必填） -->
            <div class="form-control lg:col-span-2">
              <label class="label">
                <span class="label-text font-normal">回链页面地址 <span class="text-error">*</span></span>
              </label>
              <div class="text-sm text-base-content/70 mb-2">
                💡 添加友情链接有助于提高申请通过率
              </div>
              <input 
                v-model="form.backlinkUrl" 
                type="url" 
                placeholder="https://yoursite.com/links" 
                class="input input-bordered w-full"
                :class="{ 'input-error': errors.backlinkUrl }"
                required
              />
              <label class="label">
                <span class="label-text-alt">请提供包含我们友情链接的页面地址，添加友情链接有助于提高申请通过率</span>
              </label>
              <label v-if="errors.backlinkUrl" class="label">
                <span class="label-text-alt text-error">{{ errors.backlinkUrl }}</span>
              </label>
            </div>
          </div>

          <!-- 联系信息 -->
          <div class="grid grid-cols-1 gap-6">
            <h2 class="text-xl font-normal text-base-content border-b border-base-300 pb-3 mb-2">联系信息</h2>
            
            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">联系人姓名 <span class="text-error">*</span></span>
                </label>
                <input 
                  v-model="form.contactName" 
                  type="text" 
                  placeholder="请输入您的姓名" 
                  class="input input-bordered w-full"
                  :class="{ 'input-error': errors.contactName }"
                  required
                />
                <label v-if="errors.contactName" class="label">
                  <span class="label-text-alt text-error">{{ errors.contactName }}</span>
                </label>
              </div>

              <div class="form-control">
                <label class="label">
                  <span class="label-text font-normal">联系邮箱 <span class="text-error">*</span></span>
                </label>
                <input 
                  v-model="form.contactEmail" 
                  type="email" 
                  placeholder="your@email.com" 
                  class="input input-bordered w-full"
                  :class="{ 'input-error': errors.contactEmail }"
                  required
                />
                <label v-if="errors.contactEmail" class="label">
                  <span class="label-text-alt text-error">{{ errors.contactEmail }}</span>
                </label>
              </div>
            </div>
          </div>



          <!-- 附加信息 -->
          <div class="grid grid-cols-1 gap-6">
            <h2 class="text-xl font-normal text-base-content border-b border-base-300 pb-3 mb-2">附加信息</h2>
            
            <div class="form-control">
              <label class="label">
                <span class="label-text font-normal">备注说明</span>
                <span class="label-text-alt">{{ form.remarks?.length || 0 }}/500</span>
              </label>
              <textarea 
                v-model="form.remarks" 
                class="textarea textarea-bordered h-24 w-full"
                placeholder="如有其他需要说明的信息，请在此填写"
                maxlength="500"
              ></textarea>
              <label class="label">
                <span class="label-text-alt">选填项目，可以补充说明网站特色或合作意向</span>
              </label>
            </div>
          </div>

          <!-- 提交按钮 -->
          <div class="grid grid-cols-1">
            <div class="flex justify-end pt-6 border-t border-base-300">
              <button 
                type="submit" 
                class="btn btn-primary min-w-32"
                :class="{ 'loading': isSubmitting }"
                :disabled="isSubmitting"
              >
                {{ isSubmitting ? '提交中...' : '提交申请' }}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'


// 表单数据
const form = reactive({
  siteName: '',
  siteUrl: '',
  siteLogo: '',
  category: '',
  description: '',
  contactName: '',
  contactEmail: '',
  hasBacklink: false,
  backlinkUrl: '',
  remarks: '',
  agreeTerms: false
})

// 错误信息
const errors = reactive({})

// 提交状态
const isSubmitting = ref(false)

// 表单验证
const validateForm = () => {
  // 清空之前的错误
  Object.keys(errors).forEach(key => delete errors[key])
  
  let isValid = true
  
  // 网站名称验证
  if (!form.siteName.trim()) {
    errors.siteName = '请输入网站名称'
    isValid = false
  } else if (form.siteName.length < 2 || form.siteName.length > 50) {
    errors.siteName = '网站名称长度应在2-50字符之间'
    isValid = false
  }
  
  // 网站地址验证
  if (!form.siteUrl.trim()) {
    errors.siteUrl = '请输入网站地址'
    isValid = false
  } else {
    try {
      new URL(form.siteUrl)
    } catch {
      errors.siteUrl = '请输入有效的网站地址'
      isValid = false
    }
  }
  
  // Logo地址验证（可选）
  if (form.siteLogo.trim()) {
    try {
      new URL(form.siteLogo)
    } catch {
      errors.siteLogo = '请输入有效的Logo地址'
      isValid = false
    }
  }
  
  // 分类验证
  if (!form.category) {
    errors.category = '请选择网站分类'
    isValid = false
  }
  
  // 描述验证
  if (!form.description.trim()) {
    errors.description = '请输入网站描述'
    isValid = false
  } else if (form.description.length < 10 || form.description.length > 200) {
    errors.description = '网站描述长度应在10-200字符之间'
    isValid = false
  }
  
  // 联系人姓名验证
  if (!form.contactName.trim()) {
    errors.contactName = '请输入联系人姓名'
    isValid = false
  } else if (form.contactName.length < 2 || form.contactName.length > 20) {
    errors.contactName = '联系人姓名长度应在2-20字符之间'
    isValid = false
  }
  
  // 联系邮箱验证
  if (!form.contactEmail.trim()) {
    errors.contactEmail = '请输入联系邮箱'
    isValid = false
  } else {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    if (!emailRegex.test(form.contactEmail)) {
      errors.contactEmail = '请输入有效的邮箱地址'
      isValid = false
    }
  }
  
  return isValid
}

// 提交申请
const submitApplication = async () => {
  if (!validateForm()) {
    return
  }
  
  if (!form.agreeTerms) {
    alert('请先同意友情链接申请条款')
    return
  }
  
  isSubmitting.value = true
  
  try {
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    // 显示成功消息
    alert('申请提交成功！我们会在3-7个工作日内审核您的申请，请耐心等待。')
    
    // 重置表单
    Object.keys(form).forEach(key => {
      if (typeof form[key] === 'boolean') {
        form[key] = false
      } else {
        form[key] = ''
      }
    })
    
    // 跳转回友情链接页面
    await navigateTo('/links')
    
  } catch (error) {
    console.error('提交失败:', error)
    alert('提交失败，请稍后重试')
  } finally {
    isSubmitting.value = false
  }
}
</script>
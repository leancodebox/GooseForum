<script setup lang="ts">
import { reactive } from 'vue'

// 隐私设置
const privacySettings = reactive({
  showArticles: true,
  showFollowing: true,
  emailNotifications: true
})

// 保存隐私设置
const savePrivacySettings = async () => {
  try {
    console.log('保存隐私设置:', privacySettings)
    // 这里应该调用实际的API
    alert('隐私设置保存成功')
  } catch (error) {
    console.error('保存失败:', error)
    alert('保存失败，请重试')
  }
}
</script>

<template>
  <div>
    <h3 class="card-title text-lg mb-6 border-b border-base-300 pb-3">隐私设置</h3>
    <div class="grid grid-cols-1 gap-6">
      <div class="form-control">
        <label class="label cursor-pointer justify-start gap-4">
          <input v-model="privacySettings.showArticles" type="checkbox" class="toggle toggle-primary" />
          <div>
            <span class="label-text font-normal">公开文章列表</span>
            <div class="text-sm text-base-content/60">允许其他用户查看我发布的文章</div>
          </div>
        </label>
      </div>

      <div class="form-control">
        <label class="label cursor-pointer justify-start gap-4">
          <input v-model="privacySettings.showFollowing" type="checkbox" class="toggle toggle-primary" />
          <div>
            <span class="label-text font-normal">公开关注列表</span>
            <div class="text-sm text-base-content/60">允许其他用户查看我的关注和粉丝</div>
          </div>
        </label>
      </div>

      <div class="form-control">
        <label class="label cursor-pointer justify-start gap-4">
          <input v-model="privacySettings.emailNotifications" type="checkbox" class="toggle toggle-primary" />
          <div>
            <span class="label-text font-normal">邮件通知</span>
            <div class="text-sm text-base-content/60">接收评论、点赞等活动的邮件提醒</div>
          </div>
        </label>
      </div>

      <div class="flex justify-end">
        <button @click="savePrivacySettings" class="btn btn-primary">保存隐私设置</button>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
<template>
  <div class="space-y-2">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-normal text-base-content">站点设置</h1>
        <p class="text-base-content/70 mt-1">配置站点基本信息、邮箱设置和其他选项</p>
      </div>
    </div>

    <!-- DaisyUI Tab组件 -->
    <div role="tablist" class="tabs tabs-lift">
      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="站点信息" checked/>
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300  p-6">
        <SiteInfoTab/>
      </div>

      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="邮件设置"/>
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300  p-6">
        <MailTab/>
      </div>

      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="内容设置"/>
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300  p-6">
        <ContentTab/>
      </div>

      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="用户设置"/>
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300  p-6">
        <UserTab/>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ContentTab, MailTab, SiteInfoTab, UserTab} from './site-settings/index'
</script>

<style scoped>

</style>

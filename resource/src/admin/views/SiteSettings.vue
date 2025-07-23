<template>
  <div class="space-y-2">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-normal text-base-content">站点设置</h1>
        <p class="text-base-content/70 mt-1">配置站点基本信息、SEO设置和其他选项</p>
      </div>
    </div>

    <!-- DaisyUI Tab组件 -->
    <div role="tablist" class="tabs tabs-bordered">
      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="站点信息" checked />
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300 rounded-box p-6">
        <SiteInfoTab />
      </div>

      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="SEO设置" />
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300 rounded-box p-6">
        <SeoTab />
      </div>

      <input type="radio" name="site_settings_tabs" role="tab" class="tab" aria-label="其他设置" />
      <div role="tabpanel" class="tab-content bg-base-100 border-base-300 rounded-box p-6">
        <OtherTab />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { SiteInfoTab, SeoTab, OtherTab } from './site-settings/index'
</script>

<style scoped>

</style>

<script setup>
import {NButton, NCard, NGrid, NGridItem, NStatistic, useMessage} from 'naive-ui'
import {getUserInfo} from "@/service/request"
import {onMounted, ref} from 'vue'

const message = useMessage()
const isSmallScreen = ref(window.innerWidth < 1000)

function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 1000
}

onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

function getUserInfoAction() {
  getUserInfo().then(r => {
    message.success(JSON.stringify(r.result))
  })
}
</script>

<template>
  <n-grid :cols="isSmallScreen ? 1 : 2" :x-gap="16" :y-gap="16" class="dashboard-grid">
    <!-- 统计数据卡片 -->
    <n-grid-item>
      <n-card title="统计数据" class="dashboard-card">
        <n-grid :cols="3" :x-gap="12">
          <n-grid-item>
            <n-statistic label="文章总数" :value="99">
              <template #suffix>/ 100</template>
            </n-statistic>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="用户数量" value="1,234,123"/>
          </n-grid-item>
          <n-grid-item>
            <n-statistic label="回复量" value="1,234,123"/>
          </n-grid-item>
        </n-grid>
      </n-card>
    </n-grid-item>

    <!-- 语言包卡片 -->
    <n-grid-item>
      <n-card title="语言包" class="dashboard-card">
        <n-grid :cols="3" :x-gap="12">
          <n-grid-item>
            <div class="lang-item">简体中文</div>
          </n-grid-item>
          <n-grid-item>
            <div class="lang-item">English</div>
          </n-grid-item>
          <n-grid-item>
            <div class="lang-item">日本語</div>
          </n-grid-item>
        </n-grid>
      </n-card>
    </n-grid-item>

    <!-- 功能模块卡片 -->
    <n-grid-item>
      <n-card title="功能模块" class="dashboard-card">
        <n-grid :cols="3" :x-gap="12">
          <n-grid-item>
            <n-button @click="getUserInfoAction" type="primary" class="full-width-btn">
              获取用户信息
            </n-button>
          </n-grid-item>
          <n-grid-item>
            <n-button class="full-width-btn">
              功能按钮2
            </n-button>
          </n-grid-item>
          <n-grid-item>
            <n-button class="full-width-btn">
              功能按钮3
            </n-button>
          </n-grid-item>
        </n-grid>
      </n-card>
    </n-grid-item>
    <n-grid-item>
      <n-card>
        <p>1</p>
        <p>2</p>
        <p>3</p>
        <p>4</p>
        <p>5</p>
      </n-card>
    </n-grid-item>
  </n-grid>
</template>

<style scoped>
.dashboard-grid {
  display: grid;
}

.dashboard-grid :deep(.n-grid-item) {
  display: flex;
  min-height: 0;
}

.dashboard-card {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.dashboard-card :deep(.n-card-header) {
  flex-shrink: 0;
}

.dashboard-card :deep(.n-card__content) {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.lang-item {
  padding: 12px;
  background-color: #f5f5f5;
  border-radius: 4px;
  text-align: center;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.full-width-btn {
  width: 100%;
}

/* 移动端适配 */
@media (max-width: 1000px) {
  .dashboard-card {
    margin-bottom: 0;
  }

  .lang-item {
    padding: 8px;
  }
}
</style>

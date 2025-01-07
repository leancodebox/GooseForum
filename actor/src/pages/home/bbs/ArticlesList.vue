<script setup>
import {
  NAvatar,
  NButton,
  NCard,
  NFlex,
  NIcon,
  NList,
  NListItem,
  NTag,
  NThing,
  NPagination
} from 'naive-ui'
import {onMounted, ref} from "vue";
import {getArticlesPageApi} from "@/service/request";
import {useIsMobile, useIsTablet} from "@/utils/composables";
import {useRoute, useRouter} from 'vue-router';

const router = useRouter();
const route = useRoute();
const listData = ref([])
const listContainerRef = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

function getArticlesAction(page = 1) {
  getArticlesPageApi(page, pageSize.value).then(r => {
    listData.value = r.result.list.map(function (item) {
      return {
        id: item.id,
        topic: "",
        category: "博文",
        title: item.title,
        tag: ["文章", "技术"],
        desc: item.content,
        lastUpdateTime: item.lastUpdateTime,
        body: item.content,
        username: item.username,
        avatarUrl: item.avatarUrl,
        viewCount: item.viewCount,
        commentCount: item.commentCount
      }
    })
    total.value = r.result.total || 0

    listContainerRef.value?.scrollIntoView({
      behavior: 'instant',
      block: 'start'
    })
  })
}

function handlePageChange(page) {
  currentPage.value = page
  router.push({
    query: { ...route.query, page: page }
  })
  getArticlesAction(page)
}

onMounted(() => {
  const pageFromUrl = parseInt(route.query.page) || 1;
  currentPage.value = pageFromUrl;
  getArticlesAction(pageFromUrl)
})

const text = ref('金色传说') // 需要进行高亮的文本


const isMobile = useIsMobile()
const isTablet = useIsTablet()
</script>

<template>
  <div class="articles-container" ref="listContainerRef">
    <div class="main-content">
      <!-- 标签区域 -->
      <div class="tags-section" >
        <n-tag round>技术</n-tag>
        <n-tag round>文章</n-tag>
        <n-tag round>bn</n-tag>
      </div>

      <!-- 文章列表 -->
      <n-list>
        <n-list-item v-for="item in listData">
          <router-link 
            :to="{
              path:'articlesPage',
              query: {
                title: item.title,
                id: item.id,
                page: currentPage
              }
            }"
          >
            <n-thing>
              <template #description>
                <div class="article-item">
                  <div class="article-avatar">
                    <n-avatar
                        round
                        :size="48"
                        :src="item.avatarUrl || '/api/assets/default-avatar.png'"
                    />
                  </div>

                  <div class="article-content">
                    <n-flex align="center" class="article-header" :style="{ gap: '8px' }">
                      <n-tag :bordered="false" type="success" size="small">
                        {{ item.category }}
                      </n-tag>
                      <span v-if="item.type === 'xxx'" class="blinking-div">
                        {{ item.title }}
                      </span>
                      <span v-else>
                        {{ item.title }}
                      </span>
                      <n-tag v-for="itemTag in item.tag"
                             :bordered="false"
                             size="small"
                             v-text="itemTag"
                             round
                             :style="{ marginLeft: '4px' }">
                      </n-tag>
                    </n-flex>

                    <div class="article-stats">
                      <n-flex :style="{ gap: '12px', color: '#666' }">
                        <n-flex align="center" :style="{ gap: '4px' }">
                          <n-icon size="14">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24">
                              <path fill="currentColor"
                                    d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5s5 2.24 5 5s-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3s3-1.34 3-3s-1.34-3-3-3z"/>
                            </svg>
                          </n-icon>
                          <span style="font-size: 0.9em;">{{ item.viewCount || 0 }}</span>
                        </n-flex>
                        <n-flex align="center" :style="{ gap: '4px' }">
                          <n-icon size="14">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24">
                              <path fill="currentColor"
                                    d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"/>
                            </svg>
                          </n-icon>
                          <span style="font-size: 0.9em;">{{ item.commentCount || 0 }}</span>
                        </n-flex>
                        <span style="font-size: 0.9em;">{{ item.lastUpdateTime }}</span>
                      </n-flex>
                    </div>
                  </div>
                </div>
              </template>
            </n-thing>
          </router-link>
        </n-list-item>

        <!-- 分页 -->
        <n-list-item>
          <div class="pagination-wrapper">
            <n-pagination
                v-model:page="currentPage"
                v-model:page-size="pageSize"
                :page-count="Math.ceil(total / pageSize)"
                :on-update:page="handlePageChange"
            />
          </div>
        </n-list-item>
      </n-list>
    </div>

    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="sidebar-content">
        <n-card title="统计信息">
          <p>帖子数量:0</p>
          <p>内容数量:0</p>
          <p>回复数量:0</p>
        </n-card>
        <n-card title="社区推荐" style="margin-top: 24px;">
          <div class="community-links">
            <a href="https://learnku.com/" style="text-decoration: none;">LearnKu</a>
            <a href="https://learnku.com/" style="text-decoration: none;">LearnKu</a>
            <a href="https://ruby-china.org" style="text-decoration: none;">
              <b style="color: #EB5424 !important;">Ruby</b> China
            </a>
          </div>
        </n-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.articles-container {
  display: flex;
  gap: 24px;
  padding: 24px;
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
}

.main-content {
  flex: 1;
  min-width: 0;
}

.sidebar {
  width: 360px;
  flex-shrink: 0;
  margin-bottom: 100px;
}

.sidebar-content {
  position: sticky;
  top: 80px;
}

/* 响应式布局 */
@media (max-width: 800px) {
  .articles-container {
    flex-direction: column;
  }

  .sidebar {
    width: 100%;
    margin-bottom: 0;
  }

  .sidebar-content {
    position: static;
    display: flex;
    gap: 24px;
    flex-wrap: wrap;
  }

  .sidebar-content :deep(.n-card) {
    flex: 1;
    min-width: 300px;
  }

  /* 移除第二个卡片的顶部边距 */
  .sidebar-content :deep(.n-card + .n-card) {
    margin-top: 0 !important;
  }
}

.community-links {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.tags-section {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  width: 100%;
  padding: 16px 0;
}

/* 动画相关样式保持不变 */
@keyframes highlight {
  from { background-color: #fff; }
  to { background-color: #ffff00; }
}

@keyframes yellowToRed {
  0% {
    background-color: yellow;
    color: black;
  }
  50% {
    background-color: white;
    color: red;
  }
  100% {
    background-color: yellow;
    color: black;
  }
}

.blinking-div {
  animation: yellowToRed 0.5s infinite;
}

.article-item {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.article-avatar {
  flex-shrink: 0;
}

.article-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.article-header {
  flex-wrap: wrap;
  row-gap: 8px;
}

.article-stats {
  padding-left: 0;
}

/* 响应式布局 */
@media (max-width: 800px) {
  .article-stats {
    padding-left: 0;
  }

  .article-item {
    gap: 12px;
  }

  .article-avatar {
    margin-left: 12px;
  }
}
</style>

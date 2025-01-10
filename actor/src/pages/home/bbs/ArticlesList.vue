<script setup>
import {NStatistic,NAvatar, NCard, NFlex, NIcon, NList, NListItem, NPagination, NTag, NThing} from 'naive-ui'
import {onMounted, ref} from "vue";
import {getArticlesPageApi, gtSiteStatistics} from "@/service/request";
import {useIsMobile, useIsTablet} from "@/utils/composables";
import {useRoute, useRouter} from 'vue-router';

const router = useRouter();
const route = useRoute();
const listData = ref([])
const listContainerRef = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const siteStatistic = ref({
  articleCount: 0,
  reply: 0,
  userCount: 0,
})

// 新增：分类状态
const categories = ref(['全部', '技术', '文章', 'bn']) // 示例分类
const selectedCategories = ref(['全部'])

// 新增：选择分类的处理函数
function selectCategory(category) {
    if (category === '全部') {
        if (!selectedCategories.value.includes('全部')) {
            selectedCategories.value = ['全部']
        }
    } else {
        const index = selectedCategories.value.indexOf(category)
        if (index > -1) {
            selectedCategories.value.splice(index, 1)
        } else {
            selectedCategories.value.push(category)
            // 移除 '全部' 以确保互斥
            const allIndex = selectedCategories.value.indexOf('全部')
            if (allIndex > -1) {
                selectedCategories.value.splice(allIndex, 1)
            }
        }
    }
    // 此处可添加根据选择的分类过滤文章的逻辑
}

function getArticlesAction(page = 1) {
  getArticlesPageApi(page, pageSize.value).then(r => {
    listData.value = r.result.list.map(function (item) {
      return {
        id: item.id,
        topic: "",
        category: "分享",
        title: item.title,
        tag: ["文章", "coding"],
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
    query: {...route.query, page: page}
  })
  getArticlesAction(page)
}

onMounted(async () => {
  const pageFromUrl = parseInt(route.query.page) || 1;
  currentPage.value = pageFromUrl;
  getArticlesAction(pageFromUrl)
  let resp = await gtSiteStatistics()
  siteStatistic.value = resp.result
})

const text = ref('金色传说') // 需要进行高亮的文本


const isMobile = useIsMobile()
const isTablet = useIsTablet()
</script>

<template>
  <div class="articles-container" ref="listContainerRef">
    <div class="main-content">
      <!-- 新增：分类选择区域 -->
      <div class="categories-section">
        <n-tag
          v-for="category in categories"
          :key="category"
          round
          :type="selectedCategories.includes(category) ? 'success' : 'default'"
          @click="selectCategory(category)"
          style="cursor: pointer;"
        >
          {{ category }}
        </n-tag>
      </div>

      <!-- 文章列表 -->
      <n-list>
        <n-list-item v-for="item in listData">
          <router-link
              :to="{
              path:'articlesDetail',
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
                        :size="42"
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
          <n-statistic label="帖子数量" :value="siteStatistic.articleCount"/>
          <n-statistic label="回复数量" :value="siteStatistic.reply"/>
          <n-statistic label="用户数量" :value="siteStatistic.userCount"/>
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

/* 新增：分类选择区域样式 */
.categories-section {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.categories-section .n-tag {
  transition: background-color 0.3s;
}

.categories-section .n-tag:hover {
  background-color: #f0f0f0;
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


.pagination-wrapper {
  display: flex;
  justify-content: center;
  width: 100%;
  padding: 16px 0;
}

/* 动画相关样式保持不变 */
@keyframes highlight {
  from {
    background-color: #fff;
  }
  to {
    background-color: #ffff00;
  }
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

/* 可选：根据需要调整标签的选中样式 */
.categories-section .n-tag.success {
  background-color: #52c41a; /* 绿色表示选中 */
  color: white;
}
</style>

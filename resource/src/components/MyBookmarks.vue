<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="flex justify-between items-center px-4 py-3 border-b border-base-300">
      <h2 class="text-2xl font-normal">我的收藏</h2>
    </div>
    <div class="card-body p-0">
      <!-- 加载状态 -->
      <div v-if="loading && articles.length === 0" class="p-8 text-center">
        <div class="loading loading-spinner loading-lg text-primary"></div>
        <p class="mt-4 text-base-content/60">正在加载收藏文章...</p>
      </div>

      <!-- 空状态 -->
      <div v-else-if="!loading && articles.length === 0" class="p-8 text-center">
        <div class="text-6xl mb-4">📚</div>
        <p class="text-lg font-normal text-base-content/80 mb-2">还没有收藏任何文章</p>
        <p class="text-base-content/60">快去发现感兴趣的内容吧！</p>
      </div>

      <!-- 文章列表 -->
      <ul class="list" v-else>
        <li class="list-row hover:bg-base-200 flex items-center gap-3 px-4 py-2"
            :class="{ 'opacity-50': loading }"
            v-for="article in articles"
            :key="article.id">
          <!-- 右侧内容 -->
          <div class="flex-1">
            <!-- 标题行 -->
            <div class="flex items-center gap-2 mb-1">
              <div class="badge badge-sm badge-primary flex-shrink-0">{{article.typeStr}}</div>
              <a :href="'/post/'+article.id"
                 class="text-lg font-normal text-base-content hover:text-primary hover:underline flex-1 min-w-0">{{ article.title }}</a>
            </div>
            <!-- 用户信息行和统计信息合并为一行 -->
            <div class="flex items-center justify-between text-sm text-base-content/60">
              <div class="flex items-center flex-wrap">
                <span class="mr-3">{{ article.username }}</span>
                <span class="mr-3">{{ formatDate(article.createTime) }}</span>
                <span class="badge badge-sm badge-ghost mr-1" v-for="cItem in article.categories" :key="cItem">{{ cItem }}</span>
              </div>
              <div class="flex items-center">
                <div class="flex items-center mr-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                  </svg>
                  <span class="flex-shrink-0">{{ article.viewCount }}</span>
                </div>
                <div class="flex items-center mr-4">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"/>
                  </svg>
                  <span class="flex-shrink-0"> {{ article.commentCount }} </span>
                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>

      <!-- 分页 -->
      <div v-if="totalPages > 1" class="flex justify-center items-center gap-4 p-4 border-t border-base-300">
        <button
          @click="goToPage(currentPage - 1)"
          :disabled="currentPage === 1"
          class="btn btn-sm btn-outline"
        >
          上一页
        </button>

        <span class="text-sm text-base-content/60">
          第 {{ currentPage }} 页，共 {{ totalPages }} 页
        </span>

        <button
          @click="goToPage(currentPage + 1)"
          :disabled="currentPage === totalPages"
          class="btn btn-sm btn-outline"
        >
          下一页
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { getUserBookmarkedArticles } from '../utils/gooseForumService';
import type {ArticleListItem, UserInfo} from "@/utils/gooseForumInterfaces.ts";

// 定义props（即使不使用也要定义，避免Vue警告）
const props = defineProps<{
  userInfo?: UserInfo
}>()

// 定义emits
const emit = defineEmits<{
  'user-info-updated': []
}>()



const articles = ref<ArticleListItem[]>([]);
const loading = ref(false);
const currentPage = ref(1);
const pageSize = ref(10);
const total = ref(0);

const totalPages = computed(() => Math.ceil(total.value / pageSize.value));

const loadBookmarkedArticles = async (page: number = 1) => {
  loading.value = true;
  try {
    const resp = await getUserBookmarkedArticles(page, pageSize.value);
    articles.value = resp.result.list
    total.value = resp.result.total
    currentPage.value = page
  } catch (error) {
    console.error('获取收藏文章出错:', error);
  } finally {
    loading.value = false;
  }
};

const goToPage = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    loadBookmarkedArticles(page);
  }
};

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  const now = new Date()

  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')

  // 如果是今年，不显示年份
  if (year === now.getFullYear()) {
    return `${month}-${day} ${hour}:${minute}`
  }

  return `${year}-${month}-${day} ${hour}:${minute}`
}

onMounted(() => {
  loadBookmarkedArticles();
});
</script>

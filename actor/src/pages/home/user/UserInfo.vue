<script setup>
import {NEmpty, NList, NListItem, NPagination, NSpace, NTag} from "naive-ui"
import {onMounted, ref} from "vue"
import {getUserArticles} from "@/service/request"
import {useRoute} from "vue-router"
import {formatDate} from "@/utils/dateUtil"

const route = useRoute()
const articles = ref([])
const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0,
  pageCount: 0
})

async function loadArticles(page = 1) {
  try {
    const res = await getUserArticles(page, pagination.value.pageSize)
    articles.value = res.result.list
    pagination.value = {
      pageSize: res.result.size,
      total: res.result.total,
      pageCount: Math.ceil(res.result.total / res.result.size)
    }
  } catch (error) {
    console.error('Failed to load articles:', error)
  }
}

function handlePageChange(page) {
  loadArticles(page)
}

onMounted(() => {
  loadArticles()
})
</script>

<template>
  <div class="user-articles">
    <n-list v-if="articles.length  > 0">
      <n-list-item v-for="article in articles" :key="article.id">
        <n-space vertical>
          <router-link
              :to="{
              path:'/home/bbs/articlesDetail',
              query: {
                title: article.title,
                id: article.id
              }
            }"
              class="article-title"
          >
            {{ article.title }}
          </router-link>
          <p class="article-content">{{ article.content }}</p>
          <n-space align="center" justify="space-between">
            <n-space>
              <n-tag size="small" type="success">{{ article.category }}</n-tag>
              <n-tag
                  v-for="tag in article.tags"
                  :key="tag"
                  size="small"
                  round
              >
                {{ tag }}
              </n-tag>
            </n-space>
            <span class="article-meta">
              发表于 {{ formatDate(article.createTime) }}
              · {{ article.viewCount }}浏览
              · {{ article.commentCount }}评论
            </span>
          </n-space>
        </n-space>
      </n-list-item>
    </n-list>
    <n-empty v-else description="暂无文章"/>

    <div class="pagination" v-if="pagination.total > 0">
      <n-pagination
          :page-slot="7"
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :page-count="pagination.pageCount"
          @update:page="handlePageChange"
      />
    </div>
  </div>
</template>

<style scoped>
.user-articles {
  padding: 20px;
}

.article-title {
  font-size: 16px;
  font-weight: 500;
  color: #333;
  text-decoration: none;
}

.article-title:hover {
  color: #18a058;
}

.article-content {
  color: #666;
  margin: 8px 0;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.article-meta {
  color: #999;
  font-size: 13px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>

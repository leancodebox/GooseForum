<script setup lang="ts">
import {getUserArticles} from "@/utils/articleService.ts";
import {onMounted, ref} from "vue";
import type {ArticleListItem, UserInfo} from "@/utils/articleInterfaces.ts";

const props = defineProps<{
  userInfo: UserInfo
}>()

onMounted(() => {
  refreshUserArticles()
})
let listItem = ref<ArticleListItem[]>([])
let page = ref({
  page: 1,
  size: 10,
  total: 0,
})

async function refreshUserArticles() {
  let resp = await getUserArticles(1, 10)
  listItem.value = resp.result.list
  page.value = {
    page: resp.result.page,
    size: resp.result.size,
    total: resp.result.total,
  }
}

</script>
<template>
  <div class="card bg-base-100 shadow-sm">
    <div class="card-body p-0">
      <ul class="list">
        <li class="list-row hover:bg-base-200 flex items-center gap-3 " v-for="item in listItem">
          <!-- 左侧头像 -->
          <a class="avatar" href="">
            <div class="w-10 rounded-full">
              <img :src="props.userInfo.avatarUrl" alt=".Username"/>
            </div>
          </a>
          <!-- 右侧内容 -->
          <div class="flex-1">
            <!-- 标题行 -->
            <div class="flex items-center gap-2 mb-1">
              <div class="badge badge-sm badge-primary flex-shrink-0"></div>
              <a :href="'/post/'+item.id"
                 class="text-lg font-semibold text-base-content hover:text-primary hover:underline flex-1 min-w-0">{{
                  item.title
                }}</a>
            </div>
            <!-- 用户信息行和统计信息合并为一行 -->
            <div class="flex items-center justify-between text-sm text-base-content/60">
              <div class="flex items-center flex-wrap">
                <a :href="'/user/'+props.userInfo.userId" class="mr-3">{{ props.userInfo.username }}</a>
                <span class="mr-3">2025-05-23 09:30:17</span>
                <span class="badge badge-sm badge-ghost mr-1" v-for="cItem in item.categories">{{ cItem }}</span>
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
                  <span class="flex-shrink-0">{{ item.viewCount }}</span>
                </div>
                <div class="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 mr-1"
                       fill="none"
                       viewBox="0 0 24 24"
                       stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round"
                          stroke-width="2"
                          d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"/>
                  </svg>
                  <span class="flex-shrink-0"> {{ item.commentCount }} </span>
                </div>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </div>
  <!-- 分页 -->
  <div class="flex justify-center mt-8">
    <div class="join bg-base-100 rounded-lg shadow-sm">
      <button class="join-item btn btn-sm bg-base-100 border-base-300">«</button>
      <button class="join-item btn btn-sm bg-primary text-primary-content border-primary">1</button>
      <button class="join-item btn btn-sm bg-base-100 border-base-300">2</button>
      <button class="join-item btn btn-sm bg-base-100 border-base-300">3</button>
      <button class="join-item btn btn-sm bg-base-100 border-base-300">»</button>
    </div>
  </div>
</template>
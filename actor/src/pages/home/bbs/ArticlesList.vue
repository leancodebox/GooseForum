<script setup>
import {
  NAvatar,
  NButton,
  NCard,
  NFlex,
  NIcon,
  NLayout,
  NLayoutContent,
  NLayoutSider,
  NList,
  NListItem,
  NTag,
  NThing
} from 'naive-ui'
import {onMounted, ref} from "vue";
import {getArticlesPageApi} from "@/service/request";
import {useIsMobile, useIsTablet} from "@/utils/composables";

const listData = ref([])
const containerRef = ref(null)
let maxId = 1

function getArticlesAction() {
  getArticlesPageApi(maxId).then(r => {
    let newList = r.result.list.map(function (item) {
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
    listData.value.push(...newList)
    maxId += 1
  })
}

onMounted(() => {
  maxId = 1
  getArticlesAction()
})

function more() {
  getArticlesAction()
}


const text = ref('金色传说') // 需要进行高亮的文本


const isMobile = useIsMobile()
const isTablet = useIsTablet()
</script>
<template>
  <div class="container" ref="containerRef">
    <n-layout has-sider sider-placement="right">
      <n-layout-content content-style="padding: 24px;">
        <n-flex vertical>
          <n-flex>
            <n-tag round>技术</n-tag>
            <n-tag round>文章</n-tag>
            <n-tag round>bn</n-tag>
          </n-flex>

          <n-list>
            <n-list-item v-for="item in listData">
              <router-link :to="{path:'articlesPage',query:{title:item.title,id:item.id}}">
                <n-thing>
                  <template #description>
                    <n-flex justify="space-between" align="center" :style="{ gap: '8px' }">
                      <n-flex align="center" :style="{ gap: '8px' }">
                        <n-avatar
                            round
                            size="small"
                            :src="item.avatarUrl || '/api/assets/default-avatar.png'"
                        />
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
                    </n-flex>
                  </template>
                </n-thing>
              </router-link>
            </n-list-item>
            <n-list-item>
              <n-flex vertical>
                <n-button @click="more">
                  点击一下获取更多
                </n-button>
              </n-flex>
            </n-list-item>
          </n-list>
        </n-flex>
      </n-layout-content>
      <n-layout-sider
          :width="360"
          content-style="padding: 24px;"
          bordered
          v-show="!isTablet && !isMobile"
      >
        <n-flex vertical>
          <n-card title="统计信息">
            <p>帖子数量:0</p>
            <p>内容数量:0</p>
            <p>回复数量:0</p>
          </n-card>
          <n-card title="社区推荐">
            <n-flex vertical align="center">
              <a href="https://learnku.com/" style="text-decoration: none;">LearnKu</a>
              <a href="https://ruby-china.org" style="text-decoration: none;"><b
                  style="color: #EB5424 !important;">Ruby</b> China</a>
            </n-flex>
          </n-card>
        </n-flex>
      </n-layout-sider>
    </n-layout>
  </div>
</template>
<style scoped>


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

.container {
  width: 100%;
  height: 100%;
  overflow: auto;
}
</style>

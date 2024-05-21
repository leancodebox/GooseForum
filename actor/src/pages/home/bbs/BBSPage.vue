<script setup>
import {NButton, NCard, NLayout, NLayoutContent, NLayoutSider, NList, NListItem, NSpace, NTag, NThing} from 'naive-ui'
import {onMounted, onUnmounted, ref} from "vue";
import {getArticlesPageApi} from "@/service/request";
import {useIsMobile} from "@/utils/composables";

const listData = ref([])

let maxId = 1

function getArticlesAction() {
  getArticlesPageApi(maxId).then(r => {
    let newList = r.result.list.map(function (item) {
      return {
        id: item.id,
        topic:"",
        category:"博文",
        title: item.title,
        tag: ["文章", "技术"],
        desc: item.content,
        lastUpdateTime: item.lastUpdateTime,
        body: item.content,
        username: item.username,
        // type: "xxx",
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
const textColor = ref('#000') // 文本颜色

let interval // 定时器变量

// 定义高亮动画函数，每 200ms 切换一次文本颜色
function highlightAnimation() {
  interval = setInterval(() => {
    textColor.value = textColor.value === '#fff' ? '#E21C1CFF' : '#fff'
  }, 200)
}

// 在组件挂载时启动高亮动画
onMounted(() => {
  highlightAnimation()
})

// 在组件卸载时停止高亮动画
onUnmounted(() => {
  clearInterval(interval)
})

const isMobile = useIsMobile()
</script>
<template>
  <n-layout>
    <n-layout has-sider sider-placement="right"
              class="content"
              :style="{flex: 1, maxWidth: '1400px', margin: '0 auto', padding: '24px'}">
      <n-layout-content content-style="padding: 24px;">
        <!--          layout      -->

        <n-space>
          <n-tag type="info">技术</n-tag>
          <n-tag type="info">文章</n-tag>
          <n-tag type="info">bn</n-tag>
        </n-space>

        <n-list>
          <n-list-item v-for="item in listData">
            <router-link :to="{path:'articlesPage',query:{title:item.title,id:item.id}}">
              <n-thing>
                <template #description>
                  <n-space size="small" style="padding-top: 4px" >

                    [{{ item.username }}]
                    <n-tag :bordered="false" type="success"   size="small">
                      {{ item.category}}
                    </n-tag>
                    <span v-if="item.type === 'xxx'" class="highlight-animation"
                          :style="{ color: textColor }">

                    {{ item.title }}
                    </span>
                    <span v-else>
                    {{ item.title }}
                    </span>
                    <n-tag v-for="itemTag in item.tag" :bordered="false" type="info"
                           size="small"
                           v-text="itemTag">
                    </n-tag>
                    {{ item.lastUpdateTime }}
                  </n-space>
                </template>
              </n-thing>
            </router-link>
          </n-list-item>
          <n-list-item>
            <n-button @click="more">
              more
            </n-button>
          </n-list-item>

        </n-list>

        <!--          layout      -->

      </n-layout-content>
      <n-layout-sider
          :width="360"
          content-style="padding: 24px;"
          bordered
          v-show="!isMobile"
      >
        <n-space vertical>
          <n-card title="祖国的花朵">
            <p>祖国的花朵</p>
          </n-card>

          <n-card title="今日推荐">
            <p>祖国的花朵</p>
          </n-card>

          <n-card title="友情连接">
            <p>祖国的花朵</p>
          </n-card>
        </n-space>
      </n-layout-sider>
    </n-layout>
  </n-layout>
</template>
<style scoped>
.highlight-animation {
  animation: highlight 0.5s ease-in-out alternate infinite;
}

@keyframes highlight {
  from {
    background-color: #fff;
  }
  to {
    background-color: #ffff00;
  }
}

a {
  text-decoration: none
}

@media (max-width: 1200px) {
  .content {
    max-width: 100%;
  }
}
</style>

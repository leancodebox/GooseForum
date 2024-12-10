<script setup>
import {NAnchor,NAffix,NButton, NCard, NLayout, NLayoutContent, NLayoutSider, NList, NListItem, NSpace, NTag, NThing,NFlex,NAvatar} from 'naive-ui'
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


const isMobile = useIsMobile()
const isTablet = useIsTablet()
</script>
<template>
  <div class="container" ref="containerRef">
    <n-layout has-sider sider-placement="right" >
      <n-layout-content content-style="padding: 24px;">
        <n-flex vertical>
        <n-flex>
          <n-tag  round>技术</n-tag>
          <n-tag  round>文章</n-tag>
          <n-tag  round>bn</n-tag>
        </n-flex>


        <n-list>
          <n-list-item v-for="item in listData">
            <router-link :to="{path:'articlesPage',query:{title:item.title,id:item.id}}">
              <n-thing>
                <template #description>
                    <n-flex justify="space-between">
                      <n-flex align="center" >
                        <n-avatar
                            round
                            size="small"
                            src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
                        />
                        [{{ item.username }}]
                        <n-tag :bordered="false" type="success" size="small">
                          {{ item.category }}
                        </n-tag>

                        <span v-if="item.type === 'xxx'" class="blinking-div">
                        {{ item.title }}
                        </span>
                        <span v-else>
                        {{ item.title }}
                        </span>
                        <n-tag v-for="itemTag in item.tag" :bordered="false"
                               size="small"
                               v-text="itemTag" round >
                        </n-tag>
                      </n-flex>

                      <span>
                        {{ item.lastUpdateTime }}
                      </span>
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
<!--        <n-affix-->
<!--            :listen-to="()=>containerRef"-->
<!--            :style="{height:'100%',width:'318px'}"-->
<!--            :top="80" :trigger-top="80"-->
<!--        >-->
          <n-flex vertical>
            <n-card title="祖国的花朵">
              <p>祖国的花朵</p>
            </n-card>
            <n-card title="今日推荐">
              <p>祖国的花朵</p>
            </n-card>
            <n-card title="友情连接">
              <p>祖国的花朵</p>
            </n-card>
          </n-flex>
<!--        </n-affix>-->
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

<script setup>
import {
    NButton,
    NCard,
    NH2,
    NLayout,
    NLayoutContent,
    NLayoutSider,
    NList,
    NListItem,
    NSpace,
    NTag,
    NThing
} from 'naive-ui'
import {onMounted, onUnmounted, ref} from "vue";
import {getArticlesPageApi} from "@/service/remote";

const listData = ref([])

let maxId = 0

function getArticlesAction() {
    getArticlesPageApi(maxId).then(r => {
        let newList = r.data.result.list.map(function (item) {
            return {
                id: item.id,
                title: item.title,
                tag: ["tag1", "tag2"],
                desc: item.content,
                lastUpdateTime: item.lastUpdateTime,
                body: item.content
            }
        })
        listData.value.push(...newList)
        maxId += 1
    })
}

onMounted(() => {
    maxId = 0
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
</script>
<template>
    <n-layout>
        <n-layout has-sider sider-placement="right"
                  class="content"
                  :style="{flex: 1, maxWidth: '1400px', margin: '0 auto', padding: '24px'}">
            <n-layout-content content-style="padding: 24px;">
                <!--          layout      -->
                <n-card style="margin:0 auto">
                    <n-list>
                        <n-list-item v-for="item in listData">
                            <router-link :to="{path:'articlesPage',query:{title:item.title,id:item.id}}">
                                <n-thing>
                                    <template #description>
                                        <n-space size="small" style="padding-top: 4px">
                                <span v-if="item.type === 'xxx'" class="highlight-animation"
                                      :style="{ color: textColor }">
                                    {{ text }}
                                </span>
                                            {{ item.title }}
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
                </n-card>
                <!--          layout      -->

            </n-layout-content>
            <n-layout-sider
                    :width="360"
                    content-style="padding: 24px;"
                    bordered
            >
                <n-card>
                    <n-h2>海淀桥</n-h2>
                </n-card>

                <n-card>
                    <n-h2>海淀桥</n-h2>
                </n-card>

                <n-card>
                    <n-h2>海淀桥</n-h2>
                </n-card>
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

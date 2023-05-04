<script setup>
import {
    NButton,
    NCard,
    NGrid,
    NGridItem,
    NIcon,
    NSpace,
    NStatistic,
    useMessage
} from 'naive-ui'
import {ref} from 'vue'
import {useIsMobile, useIsSmallDesktop, useIsTablet} from "@/utils/composables";
import {getUserInfo} from "@/service/remote"
import CpTool from "@/pages/manager/tool/CpTool.vue";

const message = useMessage()
const listData = ref([
    {
        title: "title1",
        tag: ["tag1", "tag2"],
        body: "bodybodybodybodybodybodybody"
    }
])

const isMobileRef = useIsMobile()
const isTabletRef = useIsTablet()
const isSmallDesktop = useIsSmallDesktop()

const isMobile = useIsMobile()


let sessionStorageData = ref("")
let localStorageData = ref("")


function set() {
    let localTmp = localStorage.getItem("tmp")
    localStorage.setItem("tmp", (Number(localTmp) + 1).toString())
    let sessionTmp = sessionStorage.getItem("tmp")
    sessionStorage.setItem("tmp", (Number(sessionTmp) + 1).toString())
    sessionStorageData.value = JSON.stringify(sessionStorage)
    localStorageData.value = JSON.stringify(localStorage)
}

function showNew() {
    sessionStorageData.value = JSON.stringify(sessionStorage)
    localStorageData.value = JSON.stringify(localStorage)
}

function getUserInfoAction() {
    getUserInfo().then(r => {
        message.success(JSON.stringify(r.data.result))
    })
}

const mdData = ref("# h1")
</script>
<template>
    <n-space vertical>
        <n-card>
            <cp-tool></cp-tool>
        </n-card>
        <n-card>
            <n-button @click="getUserInfoAction"> 获取用户信息</n-button>
        </n-card>
        <n-card title="统计数据">
            <n-grid>
                <n-grid-item :span="12">
                    <n-statistic label="统计数据" :value="99">
                        <template #prefix>
                            <n-icon>

                            </n-icon>
                        </template>
                        <template #suffix>/ 100</template>
                    </n-statistic>
                </n-grid-item>
                <n-grid-item :span="12">
                    <n-statistic label="活跃用户">1,234,123</n-statistic>
                </n-grid-item>
            </n-grid>
        </n-card>
    </n-space>
</template>
<style>
.carousel-img {
    width: 100%;
    height: 240px;
    object-fit: cover;
}
</style>
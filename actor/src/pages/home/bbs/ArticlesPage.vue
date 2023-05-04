<script setup>
import {NCard, NH2, NLayout, NLayoutContent, NLayoutSider, NList, NListItem, NSpace, NTag, NThing} from 'naive-ui'
import {onMounted, ref} from "vue";
import {useRouter} from "vue-router";
import {getArticlesDetailApi} from "@/service/remote";
import ArticlesMdPage from "@/pages/home/bbs/ArticlesMdPage.vue";
import '@/assets/github-markdown.css'


const commentList = ref([])

const articleInfo = ref({
    title: "title1",
    tag: ["tag1", "tag2"],
    createDate: "2022-12-28 01:01:01",
    lastUpdateDate: "2022-12-28 01:01:01",
    body: ""
})

let id = 1
let maxCommentId = 0

function getArticlesDetail() {
    getArticlesDetailApi(id, maxCommentId).then(r => {
        console.log(r.data.result.articleTitle)
        if (r.data.result.articleContent !== undefined && r.data.result.articleContent !== "") {
            articleInfo.value = {
                title: r.data.result.articleTitle,
                tag: ["tag1", "tag2"],
                createDate: "2022-12-28 01:01:01",
                lastUpdateDate: "2022-12-28 01:01:01",
                body: r.data.result.articleContent
            }
        }
        let commentData = r.data.result.commentList.map(function (item) {
            return {
                username: "" + item.userId,
                content: item.content,
            }
        })
        commentList.value.push(
            ...commentData
        )
    }).catch(e => {
        console.error(e)
    })
}

onMounted(() => {
    const router = useRouter();
    id = router.currentRoute.value.query.id
    getArticlesDetail()
})

</script>
<template>

    <n-layout>
        <n-layout has-sider sider-placement="right"
                  class="content"
                  :style="{flex: 1, maxWidth: '1400px', margin: '0 auto', padding: '24px'}">
            <n-layout-content content-style="padding: 24px;">


                <n-card style="margin:0 auto">

                    <h2> {{ articleInfo.title }}</h2>

                    <n-space size="small" style="margin-bottom: 16px">
                        <n-tag v-for="itemTag in articleInfo.tag" :bordered="false" type="info" size="small"
                               v-text="itemTag">
                        </n-tag>
                    </n-space>

                    <articles-md-page :markdown="articleInfo.body"></articles-md-page>

                </n-card>
                <n-card style="margin:0 auto">
                    <n-list>
                        <n-list-item v-for="item in commentList">
                            <n-thing :title="item.userId" content-style="margin-top: 10px;">
                                <articles-md-page :markdown="item.content"></articles-md-page>
                            </n-thing>
                        </n-list-item>
                    </n-list>
                </n-card>


            </n-layout-content>
            <n-layout-sider
                    :width="360"
                    content-style="padding: 24px;"
                    bordered
            >
                <n-card>
                    <n-h2>海淀桥biubiubiu</n-h2>
                </n-card>

            </n-layout-sider>
        </n-layout>
    </n-layout>
</template>


<style>
@media (max-width: 1200px) {
    .content {
        max-width: 100%;
    }
}
</style>
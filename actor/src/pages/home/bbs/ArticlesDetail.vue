<script setup>
import {
  NAlert,
  NAvatar,
  NButton,
  NCard,
  NDivider,
  NFlex,
  NInput,
  NLayout,
  NLayoutContent,
  NLayoutSider,
  NList,
  NListItem,
  NStatistic,
  NTag,
  NThing,
  useMessage
} from 'naive-ui'
import {onMounted, ref} from "vue";
import {useRouter} from "vue-router";
import {articlesReply, getArticlesDetailApi} from "@/service/request";
import ArticlesMdPage from "@/pages/home/bbs/MarkdownShow.vue";
import '@/assets/github-markdown.css'
import {useIsMobile, useIsSmallDesktop, useIsTablet} from "@/utils/composables";

const commentList = ref([])
const articleInfo = ref({
  title: "",
  userId: 0,
  username: "",
  tag: ["文章", "技术"],
  createDate: "2022-12-28 01:01:01",
  lastUpdateDate: "2022-12-28 01:01:01",
  body: ""
})

let id = 1
let maxCommentId = 0

function getArticlesDetail() {
  getArticlesDetailApi(id, maxCommentId).then(r => {
    if (r.result.articleContent !== undefined && r.result.articleContent !== "") {
      articleInfo.value = {
        userId: r.result.userId,
        username: r.result.username,
        title: r.result.articleTitle,
        tag: ["文章", "技术"],
        createDate: "2022-12-28 01:01:01",
        lastUpdateDate: "2022-12-28 01:01:01",
        body: r.result.articleContent
      }
    }
    commentList.value = r.result.commentList.map(function (item) {
      return {
        userId: item.userId,
        username: item.username,
        content: item.content,
        createTime: new Date(item.createTime).toLocaleString(),
      }
    })
  }).catch(e => {
    console.error(e)
  })
}

onMounted(() => {
  const router = useRouter();
  id = router.currentRoute.value.query.id
  getArticlesDetail()
})

const isMobile = useIsMobile()
const isTabletRef = useIsTablet()
const isSmallDesktop = useIsSmallDesktop()
const replyData = ref({
  content: "",
  replyId: 0,
})
const lockReply = ref(false)
const message = useMessage()

async function reply() {
  if (!replyData.value.content.trim()) {
    window.$message.warning('评论内容不能为空')
    return
  }

  lockReply.value = true
  try {
    let res = await articlesReply(parseInt(id), replyData.value.content, replyData.value.replyId)
    if (res.code === 0) {
      message.success('评论成功')
      replyData.value.content = ''
      getArticlesDetail()
    }
  } catch (err) {
    console.error(err)
    message.error('评论失败')
  } finally {
    lockReply.value = false
  }
}

let containerRef = ref(null)
</script>
<template>
  <div class="container" ref="containerRef">
    <n-layout has-sider sider-placement="right">
      <n-layout-content content-style="padding: 24px;">
        <n-flex vertical>
          <n-card style="margin:0 auto">
            <h2> {{ articleInfo.title }}</h2>
            <n-flex size="small" style="margin-bottom: 16px">
              <n-tag v-for="itemTag in articleInfo.tag" :bordered="false" type="info" size="small"
                     v-text="itemTag">
              </n-tag>
            </n-flex>
            <articles-md-page :markdown="articleInfo.body"></articles-md-page>
          </n-card>

          <n-card style="margin:0 auto" title="激情评论">
            <n-list v-if="commentList && commentList.length > 0" :bordered="false" class="comment-list">
              <n-list-item v-for="item in commentList" 
                           :key="item.userId" 
                           class="comment-item">
                <n-thing :title="item.username" 
                         class="comment-content">
                  <template #header-extra>
                    <span class="comment-time">{{ item.createTime }}</span>
                  </template>
                  <articles-md-page :markdown="item.content"></articles-md-page>
                </n-thing>
              </n-list-item>
            </n-list>
            <span v-else>暂无评论</span>
          </n-card>
          <n-card>
            <n-flex vertical>
              <n-alert type="info" :bordered="false">
                讨论应以学习和精进为目的。请勿发布不友善或者负能量的内容，与人为善，比聪明更重要！
              </n-alert>
              <n-input v-model:value="replyData.content"
                       type="textarea"
                       :autosize="{minRows: 3,maxRows: 5 }"
              ></n-input>
              <n-flex justify="end" size="large">
                <n-button @click="reply">评论</n-button>
              </n-flex>
            </n-flex>
          </n-card>
        </n-flex>

      </n-layout-content>
      <n-layout-sider
          :width="360"
          content-style="padding: 24px;"
          bordered
          v-show="!(isMobile||isTabletRef)"
      >
        <!--        <n-affix-->
        <!--            :listen-to="()=>containerRef"-->
        <!--            :style="{height:'100%',width:'318px'}"-->
        <!--            :top="80" :trigger-top="80"-->
        <!--        >-->
        <n-flex vertical>
          <n-card>
            <n-flex vertical>
              <n-flex>
                <n-avatar round :size="60">
                  {{ articleInfo.username }}
                </n-avatar>
                <n-flex vertical>
                  <span>{{ articleInfo.username }}({{ articleInfo.userId }})</span>
                  <span>未填写</span>
                </n-flex>
              </n-flex>
              <n-divider style="margin: 0 auto"/>
              <n-flex justify="space-around">
                <n-statistic label="声望" :value="99">
                </n-statistic>
                <n-statistic label="文章" :value="1111">
                </n-statistic>
              </n-flex>
              <n-divider style="margin: 0 auto"/>
              <n-flex justify="space-around">
                <n-button>关注</n-button>
                <n-button>私信</n-button>
              </n-flex>
            </n-flex>
          </n-card>
          <n-card title="小卡片" size="small">
            卡片内容
          </n-card>
        </n-flex>
        <!--        </n-affix>-->
      </n-layout-sider>
    </n-layout>
  </div>
</template>
<style>

.container {
  width: 100%;
  height: 100%;
  overflow: auto;
}
</style>
<style scoped>
.comment-list :deep(.n-list-item) {
  padding: 8px 0;
  border-bottom: 1px solid #eee;
}

.comment-list :deep(.n-list-item:last-child) {
  border-bottom: none;
}

.comment-content :deep(.n-thing-main) {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.comment-content :deep(.n-thing-header) {
  margin-bottom: 4px;
}

.comment-content :deep(.n-thing-header__title) {
  font-size: 14px;
  font-weight: 500;
}

.comment-time {
  font-size: 12px;
  color: #999;
}

.comment-content :deep(.n-thing-description) {
  display: none;
}

/* 调整 markdown 内容的样式 */
.comment-content :deep(.markdown-body) {
  font-size: 14px;
  line-height: 1.6;
  margin: 0;
  padding: 0;
}

/* 移除列表项的默认padding */
.comment-list :deep(.n-list) {
  --n-padding: 0;
}
</style>

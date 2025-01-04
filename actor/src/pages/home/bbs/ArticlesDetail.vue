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
import {onMounted, ref, computed} from "vue";
import {useRouter} from "vue-router";
import {articlesReply, getArticlesDetailApi, getUserInfoShow} from "@/service/request";
import ArticlesMdPage from "@/pages/home/bbs/MarkdownShow.vue";
import '@/assets/github-markdown.css'
import {useIsMobile, useIsSmallDesktop, useIsTablet} from "@/utils/composables";
import { useUserStore } from "@/modules/user";

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

const userInfo = ref({
  username: '',
  userId: 0,
  avatar: '',
  signature: '未填写',
  articleCount: 0,
  reputation: 0,
})

async function getUserInfo(userId) {
  try {
    const res = await getUserInfoShow(userId)
    if (res.code === 0 && res.result) {
      userInfo.value = {
        username: res.result.username || '',
        userId: res.result.userId || 0,
        avatar: res.result.avatar || '',
        signature: res.result.signature || '未填写',
        articleCount: res.result.articleCount || 0,
        reputation: res.result.reputation || 0,
      }
    }
  } catch (err) {
    console.error('获取用户信息失败:', err)
  }
}

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
      getUserInfo(r.result.userId)
    }
    commentList.value = r.result.commentList.map(function (item) {
      return {
        id: item.id,
        userId: item.userId,
        username: item.username,
        content: item.content,
        createTime: new Date(item.createTime).toLocaleString(),
        replyTo: item.replyToUsername,
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
  replyTo: "",
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
    cancelReply()
    lockReply.value = false
  }
}

let containerRef = ref(null)

function handleReply(item) {
  replyData.value.replyId = item.id
  replyData.value.replyTo = item.username
  document.querySelector('textarea').scrollIntoView({behavior: 'smooth'})
}

function cancelReply() {
  replyData.value.replyId = 0
  replyData.value.replyTo = ""
}

const userStore = useUserStore()
const router = useRouter()

function handleEdit() {
  router.push({
    path: '/home/write',
    query: { id: id }
  })
}

const isAuthor = computed(() => {
  console.log(userStore.userInfo, articleInfo.value.userId)
  return userStore.userInfo.userId === articleInfo.value.userId
})
</script>
<template>
  <div class="container" ref="containerRef">
    <n-layout has-sider sider-placement="right">
      <n-layout-content content-style="padding: 24px;">
        <n-flex vertical>
          <n-card style="margin:0 auto">
            <n-flex justify="space-between" align="center" style="margin-bottom: 16px">
              <h2 style="margin: 0">{{ articleInfo.title }}</h2>
              <n-button v-if="isAuthor"
                       secondary
                       size="small"
                       @click="handleEdit">
                编辑文章
              </n-button>
            </n-flex>
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
                           :key="item.id"
                           class="comment-item">
                <n-thing :title="item.username"
                         class="comment-content">
                  <template #header-extra>
                    <n-flex>
                      <n-button text size="tiny" @click="handleReply(item)">
                        回复
                      </n-button>
                      <span class="comment-time">{{ item.createTime }}</span>
                    </n-flex>
                  </template>
                  <template v-if="item.replyTo">
                    <div class="reply-reference">
                      <span class="reply-to">回复 @{{ item.replyTo }}</span>
                    </div>
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
              <div v-if="replyData.replyTo" class="reply-info">
                <span>回复 @{{ replyData.replyTo }}</span>
                <n-button text size="tiny" @click="cancelReply">取消回复</n-button>
              </div>
              <n-input v-model:value="replyData.content"
                       type="textarea"
                       :autosize="{minRows: 3,maxRows: 5 }"
              ></n-input>
              <n-flex justify="end" size="large">
                <n-button @click="reply" :loading="lockReply">评论</n-button>
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
        <n-flex vertical>
          <n-card>
            <n-flex vertical>
              <n-flex align="center" :wrap="false">
                <n-avatar
                    round
                    :size="60"
                    :src="userInfo.avatar"
                    :fallback-src="'/avatar-default.png'"
                >
                  {{ userInfo.username?.charAt(0) }}
                </n-avatar>
                <n-flex vertical style="margin-left: 12px; flex: 1;">
                  <span class="username">{{ userInfo.username }}</span>
                  <span class="signature">{{ userInfo.signature }}</span>
                </n-flex>
              </n-flex>
              <n-divider style="margin: 16px 0"/>
              <n-flex justify="space-around">
                <n-statistic :value="userInfo.reputation">
                  <template #label>
                    <div class="stat-label">声望</div>
                  </template>
                </n-statistic>
                <n-statistic :value="userInfo.articleCount">
                  <template #label>
                    <div class="stat-label">文章</div>
                  </template>
                </n-statistic>
              </n-flex>
              <n-divider style="margin: 16px 0"/>
              <n-flex justify="space-around">
                <n-button secondary size="small">关注</n-button>
                <n-button secondary size="small">私信</n-button>
              </n-flex>
            </n-flex>
          </n-card>
        </n-flex>
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

.reply-info {
  margin: 8px 0;
  padding: 4px 8px;
  background-color: #f9f9f9;
  border-radius: 4px;
  font-size: 13px;
  color: #666;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.reply-reference {
  margin-top: 8px;
  padding: 4px 8px;
  background-color: #f9f9f9;
  border-radius: 4px;
  font-size: 12px;
  color: #666;
}

.reply-to {
  color: #2080f0;
}

.username {
  font-size: 16px;
  font-weight: 500;
  line-height: 1.5;
  color: var(--n-text-color);
}

.signature {
  font-size: 13px;
  color: var(--n-text-color-3);
  line-height: 1.5;
}

.stat-label {
  font-size: 13px;
  color: var(--n-text-color-3);
}

:deep(.n-statistic-value) {
  font-size: 20px;
  font-weight: 500;
}

.edit-button {
  margin-left: 16px;
}
</style>

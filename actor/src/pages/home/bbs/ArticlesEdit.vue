<script setup>
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import {NButton, NCard, NFlex, NForm, NFormItemGi, NGrid, NInput, NSelect, useMessage} from "naive-ui"
import {onMounted, ref} from "vue";
import {getArticleCategory, writeArticles, getArticlesOrigin} from "@/service/request";
import {useRouter, useRoute} from "vue-router"

const router = useRouter()
const route = useRoute()
const message = useMessage()

let publishLoading = ref(false)
let tags = ref([])
let options = ref([])
let typeOptions = ref([
  {
    label: "博文",
    value: 0,
  }, {
    label: "教程",
    value: 1,
  }, {
    label: "问答",
    value: 2,
  }, {
    label: "分享",
    value: 3,
  },
])

let articlesData = ref({
  id: 0,
  tags: [],
  categoryId: [],
  title: "",
  content: "",
  type: 0,
})

// 获取文章原始数据
async function getOriginData() {
  const id = route.query.id
  if (!id) return

  try {
    const res = await getArticlesOrigin(id)
    if (res.code === 0 && res.result) {
      articlesData.value = {
        ...articlesData.value,
        id: parseInt(id),
        title: res.result.articleTitle,
        content: res.result.articleContent,
        // 如果后端返回了分类和类型，也需要设置
        // categoryId: res.result.categoryId,
        // type: res.result.type,
      }
    }
  } catch (err) {
    console.error('获取文章数据失败:', err)
    message.error('获取文章数据失败')
  }
}

onMounted(async () => {
  // 获取分类选项
  try {
    let data = await getArticleCategory()
    options.value = data.result.map(item => {
      return {
        label: item.name,
        value: item.value
      }
    })
    if (options.value.length > 0) {
      articlesData.value.categoryId.push(options.value[0].value)
    }
  } catch (err) {
    console.error('获取分类失败:', err)
    message.error('获取分类失败')
  }

  // 如果有 id 参数，说明是编辑模式
  if (route.query.id) {
    await getOriginData()
  }
})

async function publish(status = 1) {
  if (!articlesData.value.title.trim()) {
    message.warning('请输入标题')
    return
  }
  if (!articlesData.value.content.trim()) {
    message.warning('请输入内容')
    return
  }

  publishLoading.value = true
  try {
    let res = await writeArticles({
      ...articlesData.value,
      type: articlesData.value.type || 0,
    })
    if (res.code === 0 && res.result > 0) {
      message.success('保存成功')
      await router.push({
        path:'/home/bbs/articlesDetail',
        query: {
          id: res.result
        }
      })
    }
  } catch (e) {
    console.error(e)
    message.error('保存失败')
  } finally {
    publishLoading.value = false
  }
}

function categorySelect(value, option) {
  if (value.length > 3) {
    articlesData.value.categoryId = value.slice(0, 3)
    message.warning('最多选择3个分类')
  }
}
</script>

<!-- 模板部分基本不变，可以添加一个标题来区分是新建还是编辑 -->
<template>
  <n-card :bordered="false" style="padding: 16px;">

    <n-form
        ref="formRef"
        inline
        :label-width="80"
        :size="'medium'"
        style="margin-top: 4px;"
    >
      <n-grid cols="24" x-gap="16" item-responsive>
        <n-form-item-gi span="0:24 860:2" label="类型" path="user.age" style="min-width:80px; margin-bottom: 8px;">
          <n-select v-model:value="articlesData.type" :options="typeOptions"/>
        </n-form-item-gi>
        <n-form-item-gi span="0:24 860:4" label="分类(不超过3个)" path="user.age" style="min-width:160px; margin-bottom: 8px;">
          <n-select multiple v-model:value="articlesData.categoryId" :options="options" @update:value="categorySelect"
                    max-tag-count="responsive"/>
        </n-form-item-gi>
        <n-form-item-gi span="0:24 860:12" label="标题" style="margin-bottom: 8px;">
          <n-input v-model:value="articlesData.title"></n-input>
        </n-form-item-gi>
        <n-form-item-gi span="0:24 860:4">
          <n-flex justify="space-around">
            <n-button attr-type="button" type="info" @click="publish(0)" :disabled="publishLoading">
              存草稿
            </n-button>
            <n-button attr-type="button" type="primary" @click="publish(1)" :disabled="publishLoading">
              发文章
            </n-button>
          </n-flex>
        </n-form-item-gi>
        <n-form-item-gi :span="24">
          <mavon-editor style="width:100%;height: 100%;min-height: 500px;max-height: 600px" v-model="articlesData.content"></mavon-editor>
        </n-form-item-gi>
      </n-grid>
    </n-form>
  </n-card>
</template>

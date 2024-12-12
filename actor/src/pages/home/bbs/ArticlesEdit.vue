<script setup>
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import {NButton, NCard, NFlex, NForm, NFormItemGi, NGrid, NInput, NSelect} from "naive-ui"
import {onMounted, ref} from "vue";
import {getArticleCategory, writeArticles} from "@/service/request";
import router from "@/route/router"

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
  tags: [],
  categoryId: [],
  title: "",
  content: "",
})
onMounted(async () => {
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
})

async function publish() {
  publishLoading.value = true
  try {
    let res = await writeArticles(articlesData.value)
    if (res.result > 0) {
      await router.push({name: "articlesPage", query: {id: res.result}})
    }
  } catch (e) {
    console.error(e)
  }
  publishLoading.value = false
}

function categorySelect(value, option) {
  console.log(value, option)
}
</script>
<template>

  <n-card :bordered="false">
    <n-form
        ref="formRef"
        inline
        :label-width="80"
        :size="'medium'"
    >
      <n-grid cols="24" x-gap="24" item-responsive>



        <n-form-item-gi span="0:24 860:2" label="类型" path="user.age" style="min-width:80px">
          <n-select v-model:value="articlesData.type" :options="typeOptions"/>
        </n-form-item-gi>
        <n-form-item-gi span="0:24 860:4" label="分类(不超过3个)" path="user.age" style="min-width:160px">
          <n-select multiple v-model:value="articlesData.categoryId" :options="options" @update:value="categorySelect"
                    max-tag-count="responsive"/>
        </n-form-item-gi>
        <n-form-item-gi span="0:24 860:12" label="标题">
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
          <mavon-editor style="width:100%;height: 100%;min-height: 400px;max-height: 600px" v-model="articlesData.content"></mavon-editor>
        </n-form-item-gi>
      </n-grid>
    </n-form>


  </n-card>
</template>

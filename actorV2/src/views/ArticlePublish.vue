<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {NButton, NInput, NSelect, useMessage} from 'naive-ui'; // 引入 Naive UI 组件
import {getArticleEnum, getArticlesOrigin, submitArticle} from '@/utils/articleService'; // 引入封装的文章发布接口
import {useRoute, useRouter} from "vue-router"
import type {ArticleInfo, ArticleResponse, EnumInfoResponse} from '@/types/articleInterfaces';
import MarkdownEdit from "@/components/MarkdownEdit.vue";

const router = useRouter()
const route = useRoute()
const message = useMessage()

const articleData = ref<ArticleInfo>({
  id: 0,
  articleContent: "",
  articleTitle: "",
  categoryId: [],
  type: 0
})

const isSubmitting = ref(false); // 用于跟踪提交状态

const categories = ref([
  {label: '博客', value: 1},
  {label: '新闻', value: 2},
  {label: '教程', value: 3}
]);

const typeList = ref([
  {label: '博客', value: 1},
  {label: '新闻', value: 2},
  {label: '教程', value: 3}
]);

const submitArticleHandler = async () => {
  if (isSubmitting.value) return; // 如果正在提交，直接返回
  isSubmitting.value = true; // 设置为正在提交状态

  try {
    const response = await submitArticle<ArticleResponse>(articleData.value);
    if (response.code !== 0) {
      message.error(response.message)
      return
    }

    // 跳转到新发布的文章地址
     // 替换为实际的服务器地址
    window.location.href = `/post/${response.result}`; // 使用 window.location.href 进行跳转

  } catch (error) {
    console.error(error);

  } finally {
    isSubmitting.value = false; // 提交完成后重置状态
  }
};

// 364
onMounted(async () => {
  // 获取分类选项
  // 如果有 id 参数，说明是编辑模式
  let enumInfo = await getArticleEnum() as unknown as EnumInfoResponse; // 使用 unknown 进行类型转换
  categories.value = enumInfo.result.category.map((item) => {
    return {
      label: item.name,
      value: item.value
    }
  })
  typeList.value = enumInfo.result.type.map((item) => {
    return {
      label: item.name,
      value: item.value
    }
  })
  if (route.query.id) {
    await getOriginData()
  }
})

// 更新 getOriginData 函数的类型
async function getOriginData() {
  const id = route.query.id;

  // 确保 id 是字符串类型
  if (Array.isArray(id)) {
    console.error('ID should be a string, but received an array:', id);
    return;
  }

  if (!id) return;

  try {
    const res = await getArticlesOrigin(id) as unknown as ArticleResponse; // 使用 unknown 进行类型转换
    if (res.code === 0 && res.result) {
      articleData.value.articleTitle = res.result.articleTitle;
      articleData.value.articleContent = res.result.articleContent;
      articleData.value.categoryId = res.result.categoryId;
      articleData.value.type = res.result.type;
      articleData.value.id = parseInt(id); // 这里 id 确保是字符串
    }
  } catch (err) {
    console.error('获取文章数据失败:', err);
  }
}

</script>
<template>
  <div class="article-publish">
    <h1>发布文章</h1>
    <form @submit.prevent="submitArticleHandler" class="form">
      <div class="form-group">
        <label for="title">标题:</label>
        <n-input v-model:value="articleData.articleTitle" required placeholder="请输入标题"/>
      </div>
      <div class="form-group">
        <label for="type">类型:</label>
        <n-select
            v-model:value="articleData.type"
            :options="typeList"
            required
        />
      </div>
      <div class="form-group">
        <label for="categories">分类:</label>
        <n-select
            v-model:value="articleData.categoryId"
            :options="categories"
            multiple
            required
        />
      </div>
      <div class="form-group">
        <label for="content">内容:</label>

        <markdown-edit style="min-height: 600px; max-height: 600px;  z-index: 0;"
                      v-model="articleData.articleContent"
                      :ishljs="true"
                      required></markdown-edit>
<!--        <markdown-edit-toast  style="min-height: 600px; max-height: 600px;  z-index: 0;"-->
<!--                              v-model="articleData.articleContent"-->
<!--                              :ishljs="true"-->
<!--                              required></markdown-edit-toast>-->
      </div>
      <n-button :type="'default'" class="submit-button" @click="submitArticleHandler" :disabled="isSubmitting">发布</n-button>
    </form>
  </div>
</template>

<style scoped>
.article-publish {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.form {
  display: flex;
  flex-direction: column;
}

.form-group {
  margin-bottom: 15px;
}

label {
  margin-bottom: 5px;
  font-weight: bold;
}

.submit-button {
  margin-top: 20px;
}
</style>

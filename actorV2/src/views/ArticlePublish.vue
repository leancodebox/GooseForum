<script setup lang="ts">
import {onMounted, ref} from 'vue';
import {NButton, NInput, NSelect, useMessage} from 'naive-ui'; // 引入 Naive UI 组件
import {getArticleEnum, getArticlesOrigin, submitArticle} from '@/utils/articleService'; // 引入封装的文章发布接口
import {useRoute, useRouter} from "vue-router"
import type {ArticleInfo, ArticleResponse, EnumInfoResponse} from '@/types/articleInterfaces';
import { MdEditor } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';

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


const text = ref('# Hello Editor');
</script>
<template>
  <div class="article-publish">
    <form @submit.prevent="submitArticleHandler" class="form">
      <div class="form-group">
        <label for="title">标题:</label>
        <n-input v-model:value="articleData.articleTitle" required placeholder="请输入标题" class="input-field"/>
      </div>
      <div class="form-group inline-group">
        <div class="inline-item">
          <label for="type">类型:</label>
          <n-select
              v-model:value="articleData.type"
              :options="typeList"
              required
              class="select-field"
          />
        </div>
        <div class="inline-item">
          <label for="categories">分类:</label>
          <n-select
              v-model:value="articleData.categoryId"
              :options="categories"
              multiple
              required
              class="select-field"
          />
        </div>
      </div>
      <div class="form-group" >
        <label for="content">内容:</label>
        <MdEditor v-model="articleData.articleContent" />
      </div>
      <n-button :type="'default'" class="submit-button" @click="submitArticleHandler" :disabled="isSubmitting">发布</n-button>
    </form>
  </div>
</template>

<style scoped>
.article-publish {
  display: flex;
  flex-direction: column;
  height: 100vh; /* 使容器高度填满视口 */
}

.form {
  flex: 1; /* 使表单区域填充剩余空间 */
  display: flex;
  flex-direction: column;
  padding: 20px; /* 添加内边距 */
}

.form-group {
  margin-bottom: 10px; /* 减少底部边距 */
}

.inline-group {
  display: flex; /* 使用 Flexbox 布局 */
  gap: 10px; /* 设置子项之间的间距 */
}

.inline-item {
  flex: 1; /* 使每个子项均匀分配空间 */
}

label {
  margin-bottom: 3px; /* 减少标签底部边距 */
  font-weight: bold;
}

.input-field, .select-field {
  height: 36px; /* 设置输入框和选择框的高度 */
}

.submit-button {
  margin-top: 10px; /* 减少按钮顶部边距 */
  align-self: flex-end; /* 将按钮对齐到右侧 */
}
</style>

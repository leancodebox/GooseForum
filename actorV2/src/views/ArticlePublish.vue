<script setup lang="ts">
import {ref,onMounted} from 'vue';
import {NButton, NInput, NSelect} from 'naive-ui'; // 引入 Naive UI 组件
import {mavonEditor} from 'mavon-editor';
import 'mavon-editor/dist/css/index.css';
import {getArticlesOrigin, submitArticle} from '@/utils/articleService'; // 引入封装的文章发布接口
import {useRoute, useRouter} from "vue-router"

// 在文件顶部添加接口定义
interface ArticleResponse {
  code: number;
  result: {
    articleContent: string;
    articleTitle: string;
    categoryId: number[];
    type: string;
  };
}

const router = useRouter()
const route = useRoute()
const title = ref<string>('');
const type = ref<string>('');
const selectedCategories = ref<number[]>([]);
const content = ref<string>('');
const categories = [
  {label: '博客', value: 'blog'},
  {label: '新闻', value: 'news'},
  {label: '教程', value: 'tutorial'}
];

const submitArticleHandler = async () => {
  const article = {
    id: 0, // 示例 ID，您可以根据需要生成或获取
    title: title.value,
    type: type.value,
    categories: selectedCategories.value,
    content: content.value,
  };

  try {
    const response = await submitArticle(article);
    console.log('文章提交成功:', response);
    // 清空表单
    title.value = '';
    type.value = '';
    selectedCategories.value = [];
    content.value = '';
  } catch (error) {
    console.error(error);
  }
};

// 364
onMounted(async () => {
  // 获取分类选项
  // 如果有 id 参数，说明是编辑模式
  if (route.query.id) {
    await getOriginData()
  }
})

// 更新 getOriginData 函数的类型
async function getOriginData() {
  const id = route.query.id;
  if (!id) return;

  try {
    const res = await getArticlesOrigin(id) as unknown as ArticleResponse; // 使用 unknown 进行类型转换
    console.log(res)
    if (res.code === 0 && res.result) {
      console.log(res.result.articleContent);
      console.log(res.result.articleTitle);
      console.log(res.result.categoryId);
      title.value = res.result.articleTitle;
      content.value = res.result.articleContent;
      selectedCategories.value = res.result.categoryId
      type.value = res.result.type
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
        <n-input v-model:value="title" required placeholder="请输入标题"/>
      </div>
      <div class="form-group">
        <label for="type">类型:</label>
        <n-select
            v-model:value="type"
            :options="categories"
            required
        />
      </div>
      <div class="form-group">
        <label for="categories">分类:</label>
        <n-select
            v-model="selectedCategories"
            :options="categories"
            multiple
            required
        />
      </div>
      <div class="form-group">
        <label for="content">内容:</label>
        <mavon-editor style="width: 100%; height: 100%; min-height: 600px; max-height: 600px;z-index: 0;"
                      v-model="content"
                      required></mavon-editor>
      </div>
      <n-button :type="'default'" class="submit-button">发布</n-button>
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

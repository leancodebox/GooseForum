<script setup lang="ts">
import { ref } from 'vue';
import { NInput, NSelect, NButton } from 'naive-ui'; // 引入 Naive UI 组件
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css';

const title = ref<string>('');
const type = ref<string>('');
const selectedCategories = ref<string[]>([]);
const content = ref<string>('');
const categories = [
  { label: '博客', value: 'blog' },
  { label: '新闻', value: 'news' },
  { label: '教程', value: 'tutorial' }
];

const submitArticle = () => {
  // 处理文章发布逻辑
  console.log({
    title: title.value,
    type: type.value,
    categories: selectedCategories.value,
    content: content.value,
  });
  // 清空表单
  title.value = '';
  type.value = '';
  selectedCategories.value = [];
  content.value = '';
};

</script>
<template>
  <div class="article-publish">
    <h1>发布文章</h1>
    <form @submit.prevent="submitArticle" class="form">
      <div class="form-group">
        <label for="title">标题:</label>
        <n-input v-model:value="title" required placeholder="请输入标题" />
      </div>
      <div class="form-group">
        <label for="type">类型:</label>
        <n-select
          v-model:value="type"
          :options="[
            { label: '博客', value: 'blog' },
            { label: '新闻', value: 'news' },
            { label: '教程', value: 'tutorial' }
          ]"
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
        <mavon-editor style="width:100%;height: 100%;min-height: 600px;max-height: 600px;z-index: 0" required></mavon-editor>
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

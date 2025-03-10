<script setup lang="ts">
import { ref } from 'vue';
import { NInput, NSelect, NButton } from 'naive-ui'; // 引入 Naive UI 组件
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css';

const title = ref<string>('');
const type = ref<string>('');
const selectedCategories = ref<string[]>([]);
const content = ref<string>('');
const categories = ['技术', '生活', '旅行', '美食'];

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
  <div>
    <h1>发布文章</h1>
    <form @submit.prevent="submitArticle">
      <div>
        <label for="title">标题:</label>
        <n-input v-model:value="title" required placeholder="请输入标题" />
      </div>
      <div>
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
      <div>
        <label for="categories">分类:</label>
        <n-select
          v-model="selectedCategories"
          :options="[
            { label: '博客', value: 'blog' },
            { label: '新闻', value: 'news' },
            { label: '教程', value: 'tutorial' }
          ]"
          multiple
          required
        />
      </div>
      <div>
        <label for="content">内容:</label>
        <mavon-editor style="width:100%;height: 100%;min-height: 600px;max-height: 600px;z-index: 0" required></mavon-editor>
      </div>
      <n-button :type="'default'">发布</n-button>
    </form>
  </div>
</template>

<style scoped>
/* 添加一些样式 */
form {
  margin-bottom: 20px;
}
</style>

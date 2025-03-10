<script setup lang="ts">
import {ref} from 'vue';
import VSelect from 'vue-select';
import {MavonEditor} from 'mavon-editor';
import 'mavon-editor/dist/css/index.css'


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
        <input type="text" v-model="title" required/>
      </div>
      <div>
        <label for="type">类型:</label>
        <select v-model="type" required>
          <option value="blog">博客</option>
          <option value="news">新闻</option>
          <option value="tutorial">教程</option>
        </select>
      </div>
      <div>
        <label for="categories">分类:</label>
        <v-select
            v-model="selectedCategories"
            :options="categories"
            multiple
            required
        />
      </div>
      <div>
        <label for="content">内容:</label>
        <mavon-editor v-model="content" required></mavon-editor>
      </div>
      <button type="submit">发布</button>
    </form>
  </div>
</template>


<style scoped>
/* 添加一些样式 */
form {
  margin-bottom: 20px;
}
</style>

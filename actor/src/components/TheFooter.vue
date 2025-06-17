<script setup lang="ts">
const footerLink = [{
  type: 'text',
  url: '/',
  text: '@2025 LeanCodeBox- 分享知识，连接世界'
}, {
  type: 'a',
  url: '/',
  text: 'Goos'
}]
</script>

<template>
  <footer>
    <template v-for="item in footerLink">
      <p v-if="item.type==='text'"> {{ item.text }}</p>
      <a v-else :href="item.url">{{ item.text }}</a>
    </template>
  </footer>
</template>

<style>
</style>

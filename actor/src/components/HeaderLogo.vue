<script setup>
import {NText,NButton} from "naive-ui";
import {useThemeStore} from "@/modules/theme";

const themeStore = useThemeStore();
</script>
<template>
  <n-text tag="div" class="ui-logo" >
    <img v-if="false" alt="" src="/quote-left.png"/>
    <span class="ui-logo-span" :style="{ color: '#ff6f61' }">G</span>
    <span class="ui-logo-span" :style="{ color:  '#6ab04c' }">o</span>
    <span class="ui-logo-span" :style="{ color: themeStore.isDarkTheme ? '#e07d2d' : '#333333' }">ose</span>
    <span class="ui-logo-span" :style="{ color:  '#ff6f61' }">F</span>
    <span class="ui-logo-span" :style="{ color:  '#6ab04c' }">o</span>
    <span class="ui-logo-span" :style="{ color: themeStore.isDarkTheme ? '#e07d2d' : '#333333' }">rum</span>
  </n-text>
</template>
<style scoped>
.ui-logo {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 22px;
  margin-right: 0px;
  height: 100%;
  white-space: nowrap;
}

.ui-logo > img {
  margin-right: 12px;
  height: 32px;
  width: 32px;
}

/* 更新文本颜色和线条粗细 */
.ui-logo-span {
  font-weight: bold; /* 增加字体粗细 */
  transition: color 0.3s; /* 添加过渡效果 */
}

</style>

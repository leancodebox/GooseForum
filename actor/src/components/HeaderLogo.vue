<script setup>
import {NText} from "naive-ui";
</script>
<template>
  <n-text tag="div" class="ui-logo">
    <img v-if="false" alt="" src="/quote-left.png"/>
    <span class="ui-logo-span" style="color: #ff7e5f;">G</span>
    <span class="ui-logo-span" style="color: #feb47b;">o</span>
    <span class="ui-logo-span" style="color: #333333;">ose</span>
    <span class="ui-logo-span" style="color: #ff7e5f;">F</span>
    <span class="ui-logo-span" style="color: #feb47b;">o</span>
    <span class="ui-logo-span" style="color: #333333;">rum</span>
  </n-text>
</template>
<style scoped>
.ui-logo {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 22px;
  margin-right: 24px;
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

.ui-logo-span:hover {
  color: #555555; /* 悬停时稍微变亮 */
}
</style>

<script setup>
import {NText} from "naive-ui";
</script>
<template>
  <n-text tag="div" class="ui-logo" >
    <img alt="" src="/quote-left.png"/>
    <span>GooseForum</span>
  </n-text>
</template>
<style  scoped>

.ui-logo {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 18px;
  margin-right: 24px;
  height: 100%;
  white-space: nowrap;
}

.ui-logo > img {
  margin-right: 12px;
  height: 32px;
  width: 32px;
}
</style>

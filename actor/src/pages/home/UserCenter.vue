<script setup>
import {NFlex, NGi, NGrid} from "naive-ui"
import Left from "@/pages/home/user/Left.vue";
import Right from "@/pages/home/user/Right.vue";
</script>
<template>
  <n-grid cols="5" style="padding: 0 10px;height: 100%">
    <n-gi span="1">
      <left></left>
    </n-gi>
    <n-gi span="3">
      <router-view></router-view>
    </n-gi>
    <n-gi span="1">
      <right></right>
    </n-gi>
  </n-grid>


</template>

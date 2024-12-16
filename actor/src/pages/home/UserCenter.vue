<script setup>
import {NCard, NFlex, NGi, NGrid} from "naive-ui"
import Left from "@/pages/home/user/Left.vue";
import Right from "@/pages/home/user/Right.vue";
import {onMounted, onUnmounted, ref} from "vue";

let isSmallScreen = ref(false)

function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 800;
}

onMounted(() => {
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})
onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize);
})
</script>
<template>
  <n-card :bordered="false">
    <n-flex :justify="isSmallScreen ? 'start' : 'center'" :align-mid="true" :vertical="isSmallScreen">
      <left></left>
      <n-card style="max-width: 800px;min-width: 600px">

        <router-view></router-view>
      </n-card>
      <right></right>
    </n-flex>
  </n-card>
</template>

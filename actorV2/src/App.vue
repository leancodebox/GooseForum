<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import TheHeader from '@/components/TheHeader.vue'
import TheFooter from '@/components/TheFooter.vue'
</script>

<template>
  <div class="app-container">
    <TheHeader />
    <main>
      <RouterView />
    </main>
    <TheFooter />
  </div>
</template>

<style>
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

main {
  flex: 1;
  max-width: 1440px;
  margin: 1rem auto;
  padding: 0 2rem;
  width: 100%;
  box-sizing: border-box;
}

/* 响应式调整 */
@media (max-width: 1200px) {
  main {
    max-width: 1200px;
    padding: 0 1.5rem;
  }
}

@media (max-width: 900px) {
  main {
    padding: 0 0.25rem;
  }
}
</style>

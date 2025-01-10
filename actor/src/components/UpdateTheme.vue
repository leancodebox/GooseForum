<script setup>

import {NButton, NIcon} from "naive-ui";
import {MoonOutline, SunnyOutline} from "@vicons/ionicons5";
import {useThemeStore} from '@/modules/theme'; // Import the theme store

const themeStore = useThemeStore(); // Use the theme store

function toggleTheme() {
  themeStore.toggleTheme(); // Call the toggleTheme action from the store
}
</script>
<template>
  <n-button @click="toggleTheme" text>
    <n-icon size="24px">
      <SunnyOutline v-if="themeStore.isDarkTheme"/>
      <MoonOutline v-else/>
    </n-icon>
  </n-button>
</template>

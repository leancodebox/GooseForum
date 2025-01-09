import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useThemeStore = defineStore('theme', () => {
  const isDarkTheme = ref(false); // Reactive property for theme state

  function toggleTheme() {
    isDarkTheme.value = !isDarkTheme.value; // Toggle the theme
  }

  return {
    isDarkTheme,
    toggleTheme,
  };
}); 
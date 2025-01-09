import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useThemeStore = defineStore('theme', () => {
  const isDarkTheme = ref(false); // Reactive property for theme state

  function toggleTheme() {
    isDarkTheme.value = !isDarkTheme.value; // Toggle the theme
    document.body.classList.toggle('dark-theme', isDarkTheme.value); // Apply the class to the body
  }

  return {
    isDarkTheme,
    toggleTheme,
  };
}); 

import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useThemeStore = defineStore('theme', () => {
  // 从localStorage读取配置，如果没有则使用默认配置
  const themeConfig = ref(JSON.parse(localStorage.getItem('themeConfig') || '{"isDarkTheme": false}'));

  // 使用配置中的isDarkTheme初始化
  const isDarkTheme = ref(themeConfig.value.isDarkTheme);

  // 初始化时应用主题
  document.body.classList.toggle('dark-theme', isDarkTheme.value);

  function toggleTheme() {
    isDarkTheme.value = !isDarkTheme.value;
    // 更新主题状态
    themeConfig.value.isDarkTheme = isDarkTheme.value;
    // 保存到localStorage
    localStorage.setItem('themeConfig', JSON.stringify(themeConfig.value));
    // 应用主题
    document.body.classList.toggle('dark-theme', isDarkTheme.value);
  }

  return {
    isDarkTheme,
    toggleTheme,
  };
});

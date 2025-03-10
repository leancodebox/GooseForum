<script setup lang="ts">
import {ref, watch, onMounted} from 'vue'
import {useRouter} from 'vue-router'

const router = useRouter()
const theme = ref('light')
const isMenuOpen = ref(false)

const toggleTheme = () => {
  theme.value = theme.value === 'light' ? 'dark' : 'light'
  document.documentElement.setAttribute('data-theme', theme.value)
  localStorage.setItem('theme', theme.value)
}

// 在组件挂载时读取 localStorage 中的主题
onMounted(() => {
  const savedTheme = localStorage.getItem('theme') || 'light'
  theme.value = savedTheme
  document.documentElement.setAttribute('data-theme', savedTheme)
})

const toggleMenu = () => {
  isMenuOpen.value = !isMenuOpen.value
}

// 监听路由变化，关闭菜单
watch(
  () => router.currentRoute.value.path,
  () => {
    isMenuOpen.value = false
  }
)
</script>

<template>
  <header>
    <div class="header-main">
      <nav>
        <a href="/" class="logo">
          <span style="color:var(--primary-color)">G</span><span
            style="color:var(--secondary-color)">o</span><span
            style="color:var(--text-color)">ose</span><span style="color:var(--primary-color)">F</span><span
            style="color:var(--secondary-color)">o</span><span style="color:var(--text-color)">rum</span>
        </a>

        <div class="nav-links" :class="{ active: isMenuOpen }">
          <a href="/">首页</a>
          <a to="/post">文章</a>
        </div>

        <button class="theme-switch" @click="toggleTheme">
          {{ theme === 'light' ? '🌙' : '☀️' }}
        </button>

        <button class="hamburger" @click="toggleMenu" aria-label="菜单">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </nav>
    </div>
  </header>
</template>

<style scoped>
.header-main {
  background-color: var(--header-bg);
  box-shadow: 0 2px 4px var(--shadow-color);
  position: relative;
  z-index: 2; /* 设置一级导航的层级高于二级导航 */
}

nav {
  display: flex;
  align-items: center;
}

.nav-links {
  display: flex;
  align-items: center;
}

.theme-switch {
  margin-left: auto;
}

.hamburger {
  display: none;
  flex-direction: column;
  justify-content: space-between;
  width: 30px;
  height: 21px;
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
  margin-left: 1rem;
}

.hamburger span {
  width: 100%;
  height: 3px;
  background-color: var(--text-color);
  transition: all 0.3s ease-in-out;
}

@media screen and (max-width: 768px) {
  .hamburger {
    display: flex;
  }

  .nav-links {
    display: none;
    width: 100%;
    position: absolute;
    top: 100%;
    left: 0;
    background-color: var(--bg-color);
    flex-direction: column;
    padding: 1rem;
    box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
  }

  .nav-links.active {
    display: flex;
  }

  .nav-links a {
    padding: 0.5rem 0;
  }
}
</style>

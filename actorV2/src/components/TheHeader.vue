<script setup lang="ts">
import {ref, watch, onMounted} from 'vue'
import {useRouter} from 'vue-router'

const router = useRouter()
const theme = ref('light')
const isMenuOpen = ref(false)
const user = ref(null) // 假设你会从某处获取用户信息

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

const handleLogout = () => {
  // 处理退出登录逻辑
  console.log('Logout clicked');
}
</script>

<template>
  <header>
    <nav class="nav-container">
      <!-- Logo -->
      <div class="nav-left">
        <a href="/" class="logo" style="letter-spacing: 0;">
          <span style="color:var(--primary-color)">G</span><span
            style="color:var(--secondary-color)">o</span><span
            style="color:var(--text-color)">ose</span><span style="color:var(--primary-color)">F</span><span
            style="color:var(--secondary-color)">o</span><span style="color:var(--text-color)">rum</span>
        </a>
        <!-- 导航链接 -->
        <div class="nav-links">
          <a href="/" class="nav-link">首页</a>
          <a href="/post" class="nav-link">文章</a>
        </div>
      </div>

      <!-- 右侧功能区 -->
      <div class="nav-right">
        <!-- 主题切换按钮 -->
        <button class="theme-switch" onclick="toggleTheme()" aria-label="切换主题">
          <span class="theme-icon">🌓</span>
        </button>

        <!-- 已登录状态 -->

        <div class="user-actions" id="userActions">
          <a href="/post-edit" class="btn btn-primary">发布</a>
          <a href="/notifications" class="notification-link">
            <span class="notification-dot"></span>📬
          </a>
          <div class="user-menu">
            <button class="user-menu-btn">
              <img src="" alt="" class="user-avatar">
              <span class="username">false</span>
            </button>
            <div class="dropdown-menu">
              <a href="/user/profile">个人主页</a>
              <a href="/user/settings">设置</a>
              <a href="#" onclick="handleLogout()">退出</a>
            </div>
          </div>
        </div>
        <!-- 未登录状态 -->
        <div class="auth-buttons" id="guestButtons">
          <a href="/login" class="btn btn-auth">登录 / 注册</a>
        </div>


        <!-- 移动端用户头像 -->

        <div class="mobile-header-avatar" v-if="false">
          <img src="" alt="" class="mobile-nav-avatar">
        </div>
        <!-- 移动端菜单按钮 -->
        <button class="mobile-menu-btn" onclick="toggleMobileMenu()">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>

      <!-- 移动端菜单 -->
      <div class="mobile-menu" id="mobileMenu">
        <!-- 移动端用户信息 -->
        <div class="mobile-user-info">
          <img src="" alt="" class="mobile-user-avatar">
          <span class="mobile-username">12312</span>
        </div>
        <a href="/" class="mobile-link">首页</a>
        <a href="/post" class="mobile-link">文章</a>
        <a href="/post-edit" class="mobile-link">发布</a>
        <a href="/notifications" class="mobile-link">消息</a>
        <a href="/user/profile" class="mobile-link">个人主页</a>
        <a href="/user/settings" class="mobile-link">设置</a>
        <a href="#" onclick="handleLogout()" class="mobile-link">退出</a>
        <a href="/login" class="mobile-link" id="mobileLoginBtn">登录</a>
        <a href="/register" class="mobile-link" id="mobileRegisterBtn">注册</a>
      </div>
    </nav>
  </header>
</template>

<style scoped>
.nav-container {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 1.5rem;
  height: 60px;
  background: var(--bg-color);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.nav-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.logo {
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  height: 100%;
  padding: 0;
  font-size: 1.5rem;
  transition: transform 0.2s ease;
}

.logo:hover {
  transform: scale(1.05);
}

.nav-links {
  display: flex;
  gap: 1.2rem;
  margin-left: 1rem;
}

.nav-link {
  color: var(--text-color);
  text-decoration: none;
  padding: 0.5rem 0;
  position: relative;
}

.nav-link:after {
  content: '';
  position: absolute;
  width: 100%;
  height: 2px;
  bottom: 0;
  left: 0;
  background-color: var(--primary-color);
  transform: scaleX(0);
  transition: transform 0.3s ease;
}

.nav-link:hover:after {
  transform: scaleX(1);
}

.nav-right {
  display: flex;
  align-items: center;
  gap: 0.8rem;
}

.btn {
  padding: 0.3rem 0.6rem;
  border-radius: 4px;
  text-decoration: none;
  transition: all 0.3s ease;
  white-space: nowrap;
}

.btn-auth {
  color: var(--primary-color);
  border: 1px solid var(--primary-color);
  background: transparent;
  transition: all 0.3s ease;
  font-size: 0.9rem;
}

.btn-auth:hover {
  background: var(--primary-color);
  color: white;
}

.btn-publish {
  background: var(--secondary-color);
  color: white;
}

.notifications {
  position: relative;
}

.notification-count {
  position: absolute;
  top: -8px;
  right: -8px;
  background: red;
  color: white;
  border-radius: 50%;
  padding: 2px 6px;
  font-size: 12px;
}

.user-menu {
  position: relative;
  cursor: pointer;
  z-index: 1000;
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
}

.dropdown-menu {
  display: none;
  position: absolute;
  right: 0;
  top: 100%;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 0.5rem 0;
  min-width: 150px;
  z-index: 200;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.dropdown-menu a {
  display: block;
  padding: 0.5rem 1rem;
  color: var(--text-color);
  text-decoration: none;
}

.dropdown-menu a:hover {
  background: var(--hover-color);
}

.user-menu:hover .dropdown-menu {
  display: block;
  opacity: 1;
  transform: translateY(0);
  visibility: visible;
}

.dropdown-menu {
  opacity: 0;
  transform: translateY(-10px);
  visibility: hidden;
  transition: all 0.2s ease-in-out;
}

.mobile-menu-btn {
  display: none;
  flex-direction: column;
  gap: 4px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
}

.mobile-menu-btn span {
  display: block;
  width: 24px;
  height: 2px;
  background: var(--text-color);
  transition: all 0.3s ease;
}

.mobile-menu {
  display: none;
  position: fixed;
  top: 60px;
  left: 0;
  right: 0;
  background: var(--bg-color);
  padding: 1rem;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.mobile-user-info {
  display: flex;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 0.5rem;
}

.mobile-user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  margin-right: 1rem;
}

.mobile-username {
  color: var(--text-color);
  font-weight: 500;
}

.mobile-link {
  display: block;
  padding: 0.75rem 1rem;
  color: var(--text-color);
  text-decoration: none;
  border-bottom: 1px solid var(--border-color);
}

.mobile-header-avatar {
  display: none;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
}

.mobile-nav-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  object-fit: cover;
}

@media (max-width: 768px) {

  .mobile-header-avatar {
    display: block;
  }

  .mobile-nav-avatar {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
  }

  .user-actions {
    display: none;
  }

  .nav-links {
    display: none;
  }

  .auth-buttons, .user-actions {
    display: none !important;
  }

  .mobile-menu-btn {
    display: flex;
  }

  .mobile-menu.active {
    display: block;
  }

}
</style>

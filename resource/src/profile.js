import {createApp} from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import Page from './profile.vue'
import MyArticles from './components/MyArticles.vue'
import MyBookmarks from './components/MyBookmarks.vue'
import AccountSettings from './components/AccountSettings.vue'
import './style.css'

// 定义路由
const routes = [
  {
    path: '/',
    children: [
      {
        path: '',
        redirect: '/articles'
      },
      {
        path: 'articles',
        name: 'articles',
        component: MyArticles
      },
      {
        path: 'bookmarks',
        name: 'bookmarks', 
        component: MyBookmarks
      },
      {
        path: 'settings',
        name: 'settings',
        component: AccountSettings
      }
    ]
  }
]

// 创建路由实例
const router = createRouter({
  history: createWebHistory('/profile'),
  routes
})

const app = createApp(Page)
app.use(router)
app.mount('#app')
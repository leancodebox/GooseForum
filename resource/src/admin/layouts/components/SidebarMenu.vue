<template>
  <ul class="menu menu-md gap-1 w-full">
    <template v-for="item in menuItems" :key="item.key">
      <!-- Leaf Item -->
      <li v-if="!item.children">
        <router-link 
          :to="item.path || '#'" 
          :class="{ 'active': isPathActive(item.path!) }"
          class="rounded-lg"
        >
          <component :is="item.icon" class="w-5 h-5" />
          {{ item.label }}
        </router-link>
      </li>
      
      <!-- Parent Item -->
      <li v-else>
        <details :open="isParentActive(item)">
          <summary class="rounded-lg">
            <component :is="item.icon" class="w-5 h-5" />
            {{ item.label }}
          </summary>
          <ul>
            <li v-for="child in item.children" :key="child.key">
              <router-link 
                :to="child.path || '#'"
                :class="{ 'active': isPathActive(child.path!) }"
                class="rounded-lg"
              >
                <component :is="child.icon" class="w-4 h-4" />
                {{ child.label }}
              </router-link>
            </li>
          </ul>
        </details>
      </li>
    </template>
  </ul>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import type { MenuItem } from '@/admin/utils/menuConfig'

defineProps<{
  menuItems: MenuItem[]
}>()

const route = useRoute()

// 路径匹配函数 - 精确匹配
const isPathActive = (itemPath: string) => {
  if (!itemPath) return false
  const currentPath = route.path.replace(/\/$/, '') || '/'
  const targetPath = itemPath.replace(/\/$/, '') || '/'
  return currentPath === targetPath
}

// 判断父级菜单是否激活（子菜单中有激活项）
const isParentActive = (parentItem: MenuItem) => {
  if (!parentItem.children) return false
  return parentItem.children.some((child: MenuItem) => child.path && isPathActive(child.path))
}
</script>

<script setup>
import {NButton, NCard, NEllipsis, NFlex, NList, NListItem, NMenu} from "naive-ui"
import {onMounted, ref, h, onUnmounted} from "vue";

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '只看未读'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '全部消息'}),
    key: '2'
  }
]
let isSmallScreen = ref(false)
function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 600;
}
onMounted(()=> {
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})
onUnmounted(()=>{
  window.removeEventListener('resize', checkScreenSize);
})
</script>
<template>
  <n-card :bordered="false">
    <n-flex :justify="isSmallScreen ? 'start' : 'center'" :align-mid="true" :vertical="isSmallScreen">
      <n-menu :options="options" class="menu-component" default-value="1" />
      <n-list class="list-component">
        <n-list-item>
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
        <n-list-item>
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
        <n-list-item>
          <p><span>故人重来</span> 回答了问题 <span>webman中如何让php文件加载一次后就常驻内存了？</span></p>
          <p>6小时前</p>
        </n-list-item>
      </n-list>
    </n-flex>
  </n-card>
</template>

<style scoped>
.menu-component {
  min-width: 180px;
  max-width: 240px;
  flex: 1; /* 让菜单在垂直布局时占据可用空间 */
}

.list-component {
  min-width: 460px;
  max-width: 900px;
  flex: 2; /* 让列表在垂直布局时占据更多空间 */
}

@media (max-width: 600px) {
  .menu-component, .list-component {
    min-width: 100%; /* 在小屏幕上，让菜单和列表都占据全部宽度 */
    max-width: none;
  }

  .n-flex.vertical {
    flex-direction: column; /* 确保在垂直模式下，元素是垂直排列的 */
  }
}
</style>

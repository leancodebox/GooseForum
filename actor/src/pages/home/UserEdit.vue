<script setup>
import {NButton, NCard, NEllipsis, NFlex, NList, NListItem, NMenu} from "naive-ui"
import {h, onMounted, onUnmounted, ref} from "vue";

let options = [
  {
    label: () =>
        h(NEllipsis, null, {default: () => '个人自料'}),
    key: '1'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '头像设置'}),
    key: '2'
  },
  {
    label: () =>
        h(NEllipsis, null, {default: () => '修改密码'}),
    key: '3'
  }
]
let isSmallScreen = ref(false)

function checkScreenSize() {
  isSmallScreen.value = window.innerWidth < 800;
}

onMounted(() => {
  checkScreenSize();
  window.addEventListener('resize', checkScreenSize);
})
onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize);
})
</script>
<template>
  <n-card :bordered="false">
    <n-flex :justify="isSmallScreen ? 'start' : 'center'" :align-mid="true" :vertical="isSmallScreen">
      <n-menu :options="options" class="menu-component" default-value="1"/>
      <n-flex vertical class="list-component">
        <n-card title="个人自料" :bordered="false">
          <n-list>
            <n-list-item>
              邮箱 test@outlook.com
            </n-list-item>
            <n-list-item>
              昵称 缝合 编辑
            </n-list-item>
            <n-list-item>
              手机号 190912121 编辑
            </n-list-item>
            <n-list-item>
              个性签名 未填写 编辑
            </n-list-item>
          </n-list>
        </n-card>
        <n-card title="头像设置" :bordered="false">
          <n-list>
            <n-space vertical>
              <n-image
                  width="100"
                  src="https://07akioni.oss-cn-beijing.aliyuncs.com/07akioni.jpeg"
              />

              <n-button>上传图片</n-button>

              支持png,jpg,jpeg,gif格式 上传文件不能超过2M
            </n-space>
          </n-list>
        </n-card>

        <n-card title="密码设置" :bordered="false">
          <n-list>
            <n-list-item>
              原始密码
              <n-input></n-input>
            </n-list-item>
            <n-list-item>
              新密码
            </n-list-item>
            <n-list-item>
              确认新密码
            </n-list-item>
          </n-list>
        </n-card>
      </n-flex>
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

@media (max-width: 800px) {
  .menu-component, .list-component {
    min-width: 100%; /* 在小屏幕上，让菜单和列表都占据全部宽度 */
    max-width: none;
  }

  .n-flex.vertical {
    flex-direction: column; /* 确保在垂直模式下，元素是垂直排列的 */
  }
}
</style>


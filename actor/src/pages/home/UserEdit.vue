<script setup>
import {NButton, NCard, NEllipsis, NGi, NGrid, NImage, NInput, NList, NListItem, NMenu, NSpace} from "naive-ui"
import {h} from "vue"

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
</script>
<template>
  <n-grid cols="5" style="padding: 0 10px;height: 100%">
    <n-gi span="1">
      <n-space>
        <n-menu :options="options" style="width: 180px" default-value="1"/>
      </n-space>
    </n-gi>
    <n-gi span="4">
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
    </n-gi>
  </n-grid>
</template>

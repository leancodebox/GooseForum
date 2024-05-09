<script setup>
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import {NCard,NButton,NInput,NForm,NFormItem} from "naive-ui"
</script>
<template>

  <n-card :bordered="false" title="撰写文章">
    <n-form
        ref="formRef"
        inline
        :label-width="80"
        :size="'small'"
    >
      <n-form-item label="栏目" path="user.age">
        <n-input value="aasas" placeholder="输入年龄" />
      </n-form-item>

      <n-form-item label="标签" path="user.age">
        <n-input value="aasas" placeholder="输入年龄" />
      </n-form-item>

      <n-form-item label="系列" path="user.age">
        <n-input value="aasas" placeholder="输入年龄" />
        <n-input value="编号  " placeholder="输入年龄" />
      </n-form-item>

      <n-form-item>
        <n-button attr-type="button" >
          存草稿
        </n-button>
      </n-form-item>
      <n-form-item>
        <n-button attr-type="button" >
          发布
        </n-button>
      </n-form-item>
    </n-form>
    <mavon-editor style="height: 100%;min-height: 600px"></mavon-editor>
  </n-card>
</template>

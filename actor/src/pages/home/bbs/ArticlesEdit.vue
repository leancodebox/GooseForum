<script setup>
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import {NButton, NCard, NForm, NFormItem, NInput, NSelect} from "naive-ui"
import {ref} from "vue";

let lu = ref(null)
let tags = ref([])
let options = [
  {
    label: "Everybody's Got Something to Hide Except Me and My Monkey",
    value: 'song0',
    disabled: true
  },
  {
    label: 'Drive My Car',
    value: 'song1'
  },
  {
    label: 'Norwegian Wood',
    value: 'song2'
  },
  {
    label: "You Won't See",
    value: 'song3',
    disabled: true
  },
  {
    label: 'Nowhere Man',
    value: 'song4'
  },
  {
    label: 'Think For Yourself',
    value: 'song5'
  },
  {
    label: 'The Word',
    value: 'song6'
  },
  {
    label: 'Michelle',
    value: 'song7',
    disabled: true
  },
  {
    label: 'What goes on',
    value: 'song8'
  },
  {
    label: 'Girl',
    value: 'song9'
  },
  {
    label: "I'm looking through you",
    value: 'song10'
  },
  {
    label: 'In My Life',
    value: 'song11'
  },
  {
    label: 'Wait',
    value: 'song12'
  }
];

</script>
<template>

  <n-card :bordered="false" title="撰写文章">
    <n-form
        ref="formRef"
        inline
        :label-width="80"
        :size="'small'"
    >
      <n-form-item label="栏目" path="user.age" style="min-width:160px">
        <n-select v-model:value="lu" :options="options" />
      </n-form-item>

      <n-form-item label="标签" path="user.age" style="min-width:160px">
        <n-select
            v-model:value="tags"
            filterable
            multiple
            tag
            placeholder="输入，按回车确认"
            :show-arrow="false"
            :show="false"
        />
      </n-form-item>

      <n-form-item label="系列" path="user.age">
        <n-input value="系列1" placeholder="选择系列"/>
        <n-input value="编号  " placeholder="输入编号"/>
      </n-form-item>

      <n-form-item>
        <n-button attr-type="button" type="info">
          存草稿
        </n-button>
      </n-form-item>
      <n-form-item>
        <n-button attr-type="button" type="primary">
          发布
        </n-button>
      </n-form-item>
    </n-form>
    <mavon-editor style="height: 100%;min-height: 600px"></mavon-editor>
  </n-card>
</template>

<script setup>
import {mavonEditor} from 'mavon-editor'
import 'mavon-editor/dist/css/index.css'
import {NButton, NCard, NForm, NFormItem, NInput, NSelect} from "naive-ui"
import {ref} from "vue";

let tags = ref([])
let options = [
  // {
  //   label: "Everybody's Got Something to Hide Except Me and My Monkey",
  //   value: 'song0',
  //   disabled: true
  // },
  {
    label: '日常',
    value: 'song1'
  },
  {
    label: '分享',
    value: 'song2'
  },
  {
    label: "技术",
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
  }
];

let articlesData = ref({
  tags:[],
  category: "",
  oData: "",
})

function publish() {
  console.log(articlesData.value)
}

</script>
<template>

  <n-card :bordered="false" title="撰写文章">
    <n-form
        ref="formRef"
        inline
        :label-width="80"
        :size="'small'"
    >
      <n-form-item label="分类" path="user.age" style="min-width:160px">
        <n-select v-model:value="articlesData.category" :options="options"/>
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
        <n-button attr-type="button" type="primary" @click="publish">
          发布
        </n-button>
      </n-form-item>
    </n-form>
    <mavon-editor style="height: 100%;min-height: 600px" v-model="articlesData.oData"></mavon-editor>
  </n-card>
</template>

<template>
  <div ref="editorContainer"></div>
</template>

<script setup>
// vue-shim.d.ts 如果替换，考虑去掉这个。
import { defineProps, toRefs, onMounted, ref } from 'vue';
import { Editor } from '@toast-ui/editor';
import '@toast-ui/editor/dist/toastui-editor.css';

const props = defineProps({
  modelValue: {
    type: String,
    required: true,
  },
});

// 透传 v-model
const { modelValue } = toRefs(props);
const editorContainer = ref(null);
let editorInstance;

onMounted(() => {
  editorInstance = new Editor({
    el: editorContainer.value,
    initialEditType: 'markdown',
    previewStyle: 'vertical',
    height: '600px',
    initialValue: modelValue.value,
    events: {
      change: () => {
        const content = editorInstance.getMarkdown();
        // 这里可以通过 emit 将内容传递回父组件
      },
    },
  });
});
</script>

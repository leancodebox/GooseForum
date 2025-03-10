<template>
  <mavon-editor
      :value="modelValue"
      v-bind="$attrs"
      @change="handleChange"
  />
</template>

<script setup>
import { ref, watch, defineProps, defineEmits } from 'vue';
import { mavonEditor } from 'mavon-editor';
import 'mavon-editor/dist/css/index.css';

const props = defineProps({
  modelValue: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(['update:modelValue']);

// 手动处理编辑器内容变化
const handleChange = (value) => {
  emit('update:modelValue', value);
};
</script>

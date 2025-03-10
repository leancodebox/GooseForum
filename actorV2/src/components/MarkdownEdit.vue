<template>
  <mavon-editor
      v-model="modelValue"
      v-bind="$attrs"
      v-on="$listeners"
  />
</template>

<script setup>
// vue-shim.d.ts 如果替换，考虑去掉这个。
import { defineProps, toRefs } from 'vue';
import {mavonEditor} from 'mavon-editor';
import 'mavon-editor/dist/css/index.css';

const props = defineProps({
  modelValue: {
    type: String,
    required: true,
  },
});

// 透传 v-model
const { modelValue } = toRefs(props);
</script>

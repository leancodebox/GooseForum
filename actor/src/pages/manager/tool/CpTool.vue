<template>
    <n-input v-model:value="inputData"/>
    <n-button id="xxx-btn">复制</n-button>
</template>

<script setup>
import Clipboard from 'clipboard';
import {NInput, NButton, useMessage} from "naive-ui"
import {ref} from "vue";

const message = useMessage();

let inputData = ref("data")

let clipboard = new Clipboard('#xxx-btn', {
    text: function () {
        return inputData.value
    }
});

clipboard.on('success', function (e) {
    message.success("复制成功" + inputData.value)
});

clipboard.on('error', function (e) {
    message.success("复制失败，请手动操作")
});

</script>
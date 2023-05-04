<script setup lang="ts">
import {NButton, NH2, NCard, NLayout, NLayoutContent, NLayoutSider, NLayoutHeader, NMenu, useMessage} from "naive-ui";
import {RouterLink} from "vue-router";
import {h, ref} from "vue";
import {getQueueLenApi, remoteService, runTSpiderMaster} from "@/service/remote.js";

const message = useMessage()


</script>

<template>
    <router-view></router-view>
</template>
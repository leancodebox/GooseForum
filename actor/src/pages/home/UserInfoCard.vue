<script setup>

import {GameController} from "@vicons/ionicons5";
import {NAvatar, NIcon, NPopover, NButton, NSpace} from "naive-ui";
</script>

<template>
    <n-popover
            placement="bottom"
            trigger="hover"
    >
        <template #trigger>
            <n-avatar
                    round
                    :size="48"
            >
                旺财
            </n-avatar>
        </template>
        <n-space vertical>
            <n-button>asd</n-button>
            <n-button>asd</n-button>
            <n-button>asd</n-button>
            <n-button>asd</n-button>
        </n-space>
    </n-popover>
</template>
<template>
    <div class="markdown-body">
        <div v-html="compiledMarkdown"></div>
    </div>
</template>

<script setup>
import MarkdownIt from 'markdown-it'
import {computed} from "vue";

const props = defineProps({
    markdown: {
        type: String,
        required: true
    }
})

const compiledMarkdown = computed(() => {
    const md = new MarkdownIt({
        html:         true,        // Enable HTML tags in source
        xhtmlOut:     false,
        breaks:       true,        // Convert '\n' in paragraphs into <br>
        linkify:      false,       // 禁用自动转换URL为链接，避免punycode警告
        typographer:  false,
        quotes: '""\'\'',
    })

    // 如果需要URL转换，可以使用自定义规则
    md.normalizeLink = function(url) {
        try {
            return url.toString();
        } catch (e) {
            return '';
        }
    };

    return md.render(props.markdown || '')
})
</script>

<style scoped>
.markdown-body {
    box-sizing: border-box;
    min-width: 200px;
    max-width: 980px;
}

@media (max-width: 767px) {
    .markdown-body {
        padding: 15px;
    }
}
</style>

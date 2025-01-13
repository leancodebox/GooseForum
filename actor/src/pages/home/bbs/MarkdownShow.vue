<template>
    <div class="markdown-body">
        <div v-html="compiledMarkdown"></div>
    </div>
</template>

<script setup>
import MarkdownIt from 'markdown-it'
import {computed, watch} from "vue";
import hljs from 'highlight.js';
import {useThemeStore} from '@/modules/theme';
import 'highlight.js/styles/github.css'; // 浅色主题
import 'highlight.js/styles/github-dark.css'; // 深色主题

const themeStore = useThemeStore();

const props = defineProps({
    markdown: {
        type: String,
        required: true
    }
})

// 监听主题变化，动态切换代码高亮样式
watch(() => themeStore.isDarkTheme, (isDark) => {
    const codeBlocks = document.querySelectorAll('pre code');
    codeBlocks.forEach((block) => {
        if (block.className) {
            hljs.highlightElement(block);
        }
    });
});

const compiledMarkdown = computed(() => {
    const md = new MarkdownIt({
        html:         true,
        xhtmlOut:     false,
        breaks:       true,
        linkify:      false,
        typographer:  false,
        quotes: '""\'\'',
        highlight: function (str, lang) {
            if (lang && hljs.getLanguage(lang)) {
                try {
                    return '<pre class="hljs"><code>' + hljs.highlight(str, { language: lang }).value + '</code></pre>';
                } catch (err) {
                    console.error(err);
                }
            }
            return '<pre class="hljs"><code>' + md.utils.escapeHtml(str) + '</code></pre>';
        }
    });

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
/* 根据主题切换 markdown 样式 */
.markdown-body {
    box-sizing: border-box;
    min-width: 200px;
    max-width: 980px;
    padding: 45px;
}

@media (max-width: 767px) {
    .markdown-body {
        padding: 15px;
    }
}

/* 深色主题样式 */
.dark-theme .hljs {
    display: block;
    background: #0d1117;
    color: #c9d1d9;
}

/* 浅色主题样式 */
.hljs {
    display: block;
    background: #ffffff;
    color: #24292e;
}
</style>

<template>
    <div class="markdown-content" :class="{ 'dark-theme': themeStore.isDarkTheme }">
        <div v-html="compiledMarkdown"></div>
    </div>
</template>

<script setup>
import MarkdownIt from 'markdown-it'
import {computed, watch} from "vue";
import hljs from 'highlight.js';
import {useThemeStore} from '@/modules/theme';
// import 'highlight.js/styles/github.css'; // 浅色主题
// import 'highlight.js/styles/github-dark.css'; // 深色主题
import 'highlight.js/styles/atom-one-dark.css';   // 深色主题

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
}, { immediate: true });

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

<style>
/* 基础布局样式 */
.markdown-content {
    box-sizing: border-box;
    width: 100%;
    padding: 45px;
    transition: all 0.3s ease;
}

@media (max-width: 767px) {
    .markdown-content {
        padding: 15px;
    }
}

/* 代码块样式优化 */
.markdown-content code,
.markdown-content tt {
    padding: .2em .4em;
    margin: 0;
    font-size: 85%;
    white-space: pre-wrap;
    border-radius: 6px;
    word-break: break-word;
}

.markdown-content code br,
.markdown-content tt br {
    display: none;
}

.markdown-content pre {
    margin-top: 0;
    margin-bottom: 16px;
    padding: 8px;
    overflow: auto;
    font-size: 85%;
    line-height: 1.45;
    border-radius: 8px;
    width: 100%;
    position: relative;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.markdown-content pre code {
    display: block;
    max-width: 100%;
    padding: 0;
    margin: 0;
    overflow-x: auto;
    line-height: inherit;
    word-wrap: normal;
    border: 0;
    border-radius: 6px;
}

.markdown-content pre > code {
    padding: 0;
    margin: 0;
    word-break: normal;
    white-space: pre;
    background: transparent;
    border: 0;
}

/* 代码字体设置 */
.markdown-content code,
.markdown-content pre {
    font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
    font-size: 12px;
}

/* 主题相关样式 */
.markdown-content.dark-theme {
    color: #c9d1d9;
}

.markdown-content.dark-theme pre {
    background: #0d1117;
    border: 1px solid #30363d;
}

.markdown-content.dark-theme code {
    background-color: rgba(110,118,129,0.4);
    color: #c9d1d9;
}

/* 浅色主题样式 */
.markdown-content pre {
    background: #ffffff;
    border: 1px solid #e1e4e8;
}

.markdown-content code {
    background-color: rgba(175,184,193,0.2);
    color: #24292e;
}

/* 代码块滚动条美化 */
.markdown-content pre::-webkit-scrollbar {
    height: 8px;
    width: 8px;
}

.markdown-content pre::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 4px;
}

.markdown-content.dark-theme pre::-webkit-scrollbar-thumb {
    background: #484848;
}

.markdown-content pre::-webkit-scrollbar-track {
    background: transparent;
}
</style>

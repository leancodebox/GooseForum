<template>
    <div >
        <div v-html="compiledMarkdown"></div>
    </div>
</template>

<script setup>
import MarkdownIt from 'markdown-it'
import {computed} from "vue";
import hljs from 'highlight.js';
// import 'highlight.js/styles/intellij-light.min.css';
import 'highlight.js/styles/atom-one-dark.min.css';
import { useThemeStore } from '@/modules/theme';

const props = defineProps({
    markdown: {
        type: String,
        required: true
    }
})

const themeStore = useThemeStore();

const compiledMarkdown = computed(() => {
    const md = new MarkdownIt({
        html:         true,
        xhtmlOut:     false,
        breaks:       true,
        linkify:      false,
        typographer:  false,
        quotes: '""\'\'',
        highlight: function (str, lang) {
            const theme = themeStore.isDarkTheme ? 'dark' : 'light';
            console.log(`Current theme: ${theme}`);
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

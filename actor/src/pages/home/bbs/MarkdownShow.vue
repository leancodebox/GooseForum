<template>
  <div class="markdown-content" :class="{ 'dark-theme': themeStore.isDarkTheme }">
    <div v-html="compiledMarkdown"></div>
  </div>
</template>

<script setup>
import MarkdownIt from 'markdown-it'
import {computed, onMounted, watch} from "vue";
import hljs from 'highlight.js';
import {useThemeStore} from '@/modules/theme';
import './github-theme.css';

const themeStore = useThemeStore();
const props = defineProps({
  content: {
    type: String,
    required: true
  }
})

// 创建 markdown-it 实例并配置
const md = new MarkdownIt({
  html: true,
  breaks: true,
  linkify: true,
  highlight: function (str, lang) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return '<pre><code class="hljs">'
            + hljs.highlight(str, {language: lang}).value
            + '</code></pre>';
      } catch (err) {}
    }
    return '<pre><code class="hljs">' + md.utils.escapeHtml(str) + '</code></pre>';
  }
});


// 监听主题变化和markdown内容变化，重新应用高亮
watch([() => themeStore.isDarkTheme, () => props.content], () => {

});

// 组件挂载时初始化高亮
onMounted(() => {

});

const compiledMarkdown = computed(() => {
  return md.render(props.content || '')
})
</script>

<style>
/* 基础布局样式 */
.markdown-content {
  box-sizing: border-box;
  width: 100%;
  padding: 24px;
  transition: all 0.3s ease;
}

/* 代码块容器样式 */
.markdown-content pre {
  margin: 16px 0;
  padding: 16px;
  border-radius: 8px;
  position: relative;
  overflow-x: auto;
  overflow-y: hidden;
  border: 1px solid #e1e4e8;
  background-color: #f6f8fa;
}

/* 行内代码样式 */
.markdown-content :not(pre) > code {
  padding: 0.2em 0.4em;
  margin: 0 0.2em;
  font-size: 85%;
  border-radius: 4px;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  white-space: break-spaces;
  word-wrap: break-word;
}

/* 代码块内容样式 */
.markdown-content pre > code.hljs {
  display: block;
  padding: 0;
  overflow: visible;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 14px;
  line-height: 1.6;
  tab-size: 4;
  white-space: pre;
  word-break: normal;
  word-wrap: normal;
  background: transparent;
}

/* 浅色主题 */
.markdown-content {

}

.markdown-content :not(pre) > code {
  background-color: rgba(175, 184, 193, 0.2);

}

/* 深色主题 */
.markdown-content.dark-theme {
  color: #c9d1d9;
}

.markdown-content.dark-theme :not(pre) > code {
  background-color: rgba(110, 118, 129, 0.4);
  color: #c9d1d9;
}

/* 滚动条样式 */
.markdown-content pre::-webkit-scrollbar {
  width: 0;
  height: 12px;
}

.markdown-content pre::-webkit-scrollbar-track {
  background-color: transparent;
}

.markdown-content pre::-webkit-scrollbar-thumb {
  background-color: rgba(69, 79, 89, 0.3);
  border-radius: 6px;
  border: 4px solid transparent;
  background-clip: content-box;
  transition: background-color .2s;
}

.markdown-content pre::-webkit-scrollbar-thumb:hover {
  background-color: rgba(69, 79, 89, 0.5);
}

/* 深色主题滚动条 */
.markdown-content.dark-theme pre::-webkit-scrollbar-thumb {
  background-color: rgba(255, 255, 255, 0.2);
}

.markdown-content.dark-theme pre::-webkit-scrollbar-thumb:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

/* 深色主题下的代码块边框 */
.markdown-content.dark-theme pre {
  border-color: #30363d;
}

.markdown-content.dark-theme pre {
  background-color: #0d1117;
}

@media (max-width: 767px) {
  .markdown-content {
    padding: 16px;
  }

  .markdown-content pre {
    padding: 12px;
    margin: 12px 0;
  }
}
</style>

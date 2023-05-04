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
        xhtmlOut:     false,        // Use '/' to close single tags (<br />).
                                    // This is only for full CommonMark compatibility.
        breaks:       true,        // Convert '\n' in paragraphs into <br>
        // langPrefix:   'language-',  // CSS language prefix for fenced blocks. Can be
                                    // useful for external highlighters.
        linkify:      true,        // Autoconvert URL-like text to links

        // Enable some language-neutral replacement + quotes beautification
        // For the full list of replacements, see https://github.com/markdown-it/markdown-it/blob/master/lib/rules_core/replacements.js
        typographer:  false,

        // Double + single quotes replacement pairs, when typographer enabled,
        // and smartquotes on. Could be either a String or an Array.
        //
        // For example, you can use '«»„“' for Russian, '„“‚‘' for German,
        // and ['«\xA0', '\xA0»', '‹\xA0', '\xA0›'] for French (including nbsp).
        // quotes: '“”‘’',

        // Highlighter function. Should return escaped HTML,
        // or '' if the source string is not changed and should be escaped externally.
        // If result starts with <pre... internal wrapper is skipped.
        // highlight: function (/*str, lang*/) { return ''; }
    })
    return md.render(props.markdown)
})

</script>

<style scoped>
.markdown-body {
    box-sizing: border-box;
    min-width: 200px;
    max-width: 980px;
    /*margin: 0 auto;*/
    /*padding: 40px;*/
}

@media (max-width: 767px) {
    .markdown-body {
        padding: 15px;
    }
}
</style>

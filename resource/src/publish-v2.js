import './style.css'

// 检查是否在Vue环境中，如果不是则直接初始化markdown编辑器
if (document.getElementById('markdown-editor')) {
    initMarkdownEditor()
}

function initMarkdownEditor() {
    const titleInput = document.getElementById('article-title')
    const markdownEditor = document.getElementById('markdown-editor')
    const previewTitle = document.getElementById('preview-title')
    const previewContent = document.getElementById('preview-content')

    // 配置marked选项
    if (typeof marked !== 'undefined') {
        marked.setOptions({
            breaks: true,
            gfm: true,
            highlight: function(code, lang) {
                return code // 简单返回代码，不做语法高亮
            }
        })
    }

    // 更新预览内容
    function updatePreview() {
        const title = titleInput.value.trim() || '文章标题预览'
        const markdown = markdownEditor.value

        // 更新标题
        previewTitle.textContent = title

        // 转换markdown到HTML
        if (markdown.trim()) {
            try {
                let html = ''
                if (typeof marked !== 'undefined') {
                    html = marked.parse(markdown)
                    // 使用DOMPurify清理HTML（如果可用）
                    if (typeof DOMPurify !== 'undefined') {
                        html = DOMPurify.sanitize(html)
                    }
                } else {
                    // 如果marked不可用，做简单的markdown转换
                    html = simpleMarkdownToHtml(markdown)
                }
                previewContent.innerHTML = html
            } catch (error) {
                console.error('Markdown解析错误:', error)
                previewContent.innerHTML = '<p class="text-error">Markdown解析出错，请检查语法</p>'
            }
        } else {
            previewContent.innerHTML = '<p class="text-base-content/60">在左侧编辑区域输入内容，预览将在这里显示...</p>'
        }
    }

    // 简单的markdown转HTML函数（备用）
    function simpleMarkdownToHtml(markdown) {
        return markdown
            .replace(/^### (.*$)/gim, '<h3>$1</h3>')
            .replace(/^## (.*$)/gim, '<h2>$1</h2>')
            .replace(/^# (.*$)/gim, '<h1>$1</h1>')
            .replace(/\*\*(.*?)\*\*/gim, '<strong>$1</strong>')
            .replace(/\*(.*?)\*/gim, '<em>$1</em>')
            .replace(/^- (.*$)/gim, '<li>$1</li>')
            .replace(/\n/gim, '<br>')
    }

    // 字符计数功能
    const charCountElement = document.getElementById('char-count')
    function updateCharCount() {
        const count = markdownEditor.value.length
        if (charCountElement) {
            charCountElement.textContent = count.toLocaleString()
        }
    }

    // 监听输入事件
    titleInput.addEventListener('input', updatePreview)
    markdownEditor.addEventListener('input', function() {
        updatePreview()
        updateCharCount()
    })

    // 按钮功能
    const clearButton = document.querySelector('.btn-ghost')
    const publishButton = document.querySelector('.btn-primary')
    
    if (clearButton) {
        clearButton.addEventListener('click', function() {
            if (confirm('确定要清空所有内容吗？此操作不可撤销。')) {
                titleInput.value = ''
                markdownEditor.value = ''
                updatePreview()
                updateCharCount()
            }
        })
    }
    
    if (publishButton) {
        publishButton.addEventListener('click', function() {
            const title = titleInput.value.trim()
            const content = markdownEditor.value.trim()
            
            if (!title) {
                alert('请输入文章标题')
                titleInput.focus()
                return
            }
            
            if (!content) {
                alert('请输入文章内容')
                markdownEditor.focus()
                return
            }
            
            // 这里可以添加实际的发布逻辑
            alert('文章发布功能待实现')
        })
    }

    // 初始化预览和字符计数
    updatePreview()
    updateCharCount()

    // 添加一些示例内容
    if (!markdownEditor.value.trim()) {
        markdownEditor.value = ``
        titleInput.value = '我的第一篇文章'
        updatePreview()
    }
}
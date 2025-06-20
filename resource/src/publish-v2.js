import './style.css'

// 导入npm安装的库
import { marked } from 'marked'
import DOMPurify from 'dompurify'

// 文章数据
const articleData = {
    id: 0,
    content: '',
    title: '',
    categoryId: [],
    type: 1
}

// 动态选项数据
const categories = []
const typeList = []

// 状态管理
let isSubmitting = false

// 检查是否在Vue环境中，如果不是则直接初始化markdown编辑器
if (document.getElementById('markdown-editor')) {
    initMarkdownEditor()
}

function initMarkdownEditor() {
    const titleInput = document.getElementById('article-title')
    const markdownEditor = document.getElementById('markdown-editor')
    const previewTitle = document.getElementById('preview-title')
    const previewContent = document.getElementById('preview-content')
    const typeSelect = document.getElementById('article-type')
    const categorySelect = document.getElementById('article-category')

    // 配置marked选项
    marked.setOptions({
        breaks: true,
        gfm: true,
        headerIds: false,
        mangle: false
    })

    // 初始化数据
    initData()

    // 更新预览内容
    window.updatePreview = function() {
        const titleInput = document.getElementById('article-title')
        const markdownEditor = document.getElementById('markdown-editor')
        const previewTitle = document.getElementById('preview-title')
        const previewContent = document.getElementById('preview-content')
        
        if (!titleInput || !markdownEditor || !previewTitle || !previewContent) {
            return
        }
        
        const title = titleInput.value.trim() || '文章标题预览'
        const markdown = markdownEditor.value

        // 更新标题
        previewTitle.textContent = title

        // 转换markdown到HTML
        if (markdown.trim()) {
            try {
                // 使用marked解析markdown
                let html = marked.parse(markdown)
                // 使用DOMPurify清理HTML
                html = DOMPurify.sanitize(html)
                previewContent.innerHTML = html
            } catch (error) {
                console.error('Markdown解析错误:', error)
                previewContent.innerHTML = '<p class="text-error">Markdown解析出错，请检查语法</p>'
            }
        } else {
            previewContent.innerHTML = '<p class="text-base-content/60">在左侧编辑区域输入内容，预览将在这里显示...</p>'
        }
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
    titleInput.addEventListener('input', window.updatePreview)
    markdownEditor.addEventListener('input', function() {
        window.updatePreview()
        updateCharCount()
    })

    // 清空内容按钮事件
    const clearButton = document.querySelector('.btn-ghost')
    if (clearButton) {
        clearButton.addEventListener('click', () => {
            if (confirm('确定要清空所有内容吗？')) {
                titleInput.value = ''
                markdownEditor.value = ''
                typeSelect.value = ''
                categorySelect.value = ''
                articleData.title = ''
                articleData.content = ''
                articleData.type = 1
                articleData.categoryId = []
                window.updatePreview()
            }
        })
    }

    // 发布按钮事件
    const publishButton = document.querySelector('#submit-article')
    if (publishButton) {
        publishButton.addEventListener('click', () => {
            submitArticle()
        })
    }

    // 表单字段事件监听
    titleInput.addEventListener('input', (e) => {
        articleData.title = e.target.value
    })

    markdownEditor.addEventListener('input', (e) => {
        articleData.content = e.target.value
    })

    typeSelect.addEventListener('change', (e) => {
        articleData.type = parseInt(e.target.value) || 1
    })

    categorySelect.addEventListener('change', (e) => {
        const selectedOptions = Array.from(e.target.selectedOptions)
        articleData.categoryId = selectedOptions.map(option => parseInt(option.value)).filter(id => !isNaN(id))
    })

    // 初始化预览和字符计数
     window.updatePreview()
     updateCharCount()
 }

// 初始化数据
async function initData() {
    try {
        // 获取分类和类型选项
        await getArticleEnum()
        
        // 检查是否为编辑模式
        const urlParams = new URLSearchParams(window.location.search)
        const articleId = urlParams.get('id')
        
        if (articleId) {
            articleData.id = parseInt(articleId)
            await getOriginData(articleId)
        }
    } catch (error) {
        console.error('初始化数据失败:', error)
        showMessage('初始化失败，请刷新页面重试', 'error')
    }
}

// 获取文章枚举数据
async function getArticleEnum() {
    try {
        const response = await fetch('/api/forum/get-articles-enum', {
            method: 'GET'
        })
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const result = await response.json()
        
        if (result.code === 0) {
            // 填充类型选项
            const typeSelect = document.getElementById('article-type')
            if (typeSelect && result.result.type) {
                typeList.length = 0
                typeList.push(...result.result.type)
                
                typeSelect.innerHTML = '<option value="">请选择类型</option>'
                result.result.type.forEach(type => {
                    const option = document.createElement('option')
                    option.value = type.id
                    option.textContent = type.name
                    typeSelect.appendChild(option)
                })
            }
            
            // 填充分类选项
            const categorySelect = document.getElementById('article-category')
            if (categorySelect && result.result.category) {
                categories.length = 0
                categories.push(...result.result.category)
                
                categorySelect.innerHTML = ''
                result.result.category.forEach(category => {
                    const option = document.createElement('option')
                    option.value = category.id
                    option.textContent = category.name
                    categorySelect.appendChild(option)
                })
            }
        } else {
            throw new Error(result.msg || '获取枚举数据失败')
        }
    } catch (error) {
        console.error('获取枚举数据失败:', error)
        throw error
    }
}

// 获取原始文章数据（编辑模式）
async function getOriginData(articleId) {
    try {
        const response = await fetch(`/api/forum/get-articles-origin`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                id: parseInt(articleId)
            })
        })
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const result = await response.json()
        
        if (result.code === 0 && result.result) {
            const data = result.result
            
            // 更新文章数据
            articleData.title = data.articleTitle || ''
            articleData.content = data.articleContent || ''
            articleData.type = data.type || 1
            articleData.categoryId = data.categoryId || []
            
            // 更新表单字段
            const titleInput = document.getElementById('article-title')
            const markdownEditor = document.getElementById('markdown-editor')
            const typeSelect = document.getElementById('article-type')
            const categorySelect = document.getElementById('article-category')
            
            if (titleInput) titleInput.value = articleData.title
            if (markdownEditor) markdownEditor.value = articleData.content
            
            // 确保类型选择正确映射
            if (typeSelect) {
                typeSelect.value = articleData.type.toString()
            }
            
            if (categorySelect) {
                Array.from(categorySelect.options).forEach(option => {
                    option.selected = articleData.categoryId.includes(parseInt(option.value))
                })
            }
            
            // 更新预览
            if (window.updatePreview) {
                window.updatePreview()
            }
        } else {
            throw new Error(result.msg || '获取文章数据失败')
        }
    } catch (error) {
        console.error('获取文章数据失败:', error)
        throw error
    }
}

// 表单验证
function validateForm() {
    if (!articleData.title.trim()) {
        showMessage('请输入文章标题', 'error')
        document.getElementById('article-title')?.focus()
        return false
    }
    
    if (!articleData.content.trim()) {
        showMessage('请输入文章内容', 'error')
        document.getElementById('markdown-editor')?.focus()
        return false
    }
    
    if (!articleData.type) {
        showMessage('请选择文章类型', 'error')
        document.getElementById('article-type')?.focus()
        return false
    }
    
    if (!articleData.categoryId.length) {
        showMessage('请选择文章分类', 'error')
        document.getElementById('article-category')?.focus()
        return false
    }
    
    return true
}

// 提交文章
async function submitArticle() {
    if (isSubmitting) return
    
    if (!validateForm()) return
    
    isSubmitting = true
    const publishButton = document.querySelector('.btn-primary')
    const originalText = publishButton?.textContent
    
    try {
        if (publishButton) {
            publishButton.disabled = true
            publishButton.textContent = '发布中...'
        }
        
        const response = await fetch('/api/forum/submit-article', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token') || ''}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(articleData)
        })
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`)
        }
        
        const result = await response.json()
        
        if (result.code === 0) {
            showMessage(articleData.id ? '文章更新成功！' : '文章发布成功！', 'success')
            
            // 延迟跳转到文章列表或详情页
            setTimeout(() => {
                window.location.href = '/'
            }, 1500)
        } else {
            throw new Error(result.msg || '提交失败')
        }
    } catch (error) {
        console.error('提交文章失败:', error)
        showMessage(error.message || '提交失败，请重试', 'error')
    } finally {
        isSubmitting = false
        if (publishButton) {
            publishButton.disabled = false
            publishButton.textContent = originalText
        }
    }
}

// 显示消息提示
function showMessage(message, type = 'info') {
    // 创建消息元素
    const messageEl = document.createElement('div')
    messageEl.className = `alert alert-${type === 'error' ? 'error' : type === 'success' ? 'success' : 'info'} fixed top-4 right-4 w-auto max-w-sm z-50`
    messageEl.innerHTML = `
        <span>${message}</span>
        <button class="btn btn-sm btn-ghost" onclick="this.parentElement.remove()">×</button>
    `
    
    document.body.appendChild(messageEl)
    
    // 自动移除
    setTimeout(() => {
        if (messageEl.parentElement) {
            messageEl.remove()
        }
    }, 5000)
}
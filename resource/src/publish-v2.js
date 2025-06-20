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

// 分类选择器配置
const categoryConfig = {
    maxSelection: 3,
    selectedCategories: new Set(),
    filteredCategories: []
}

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

    // 初始化分类选择器
    initCategorySelector()

    // 保留原有的隐藏select的change事件（用于兼容性）
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
                    option.value = type.value
                    option.textContent = type.name
                    typeSelect.appendChild(option)
                })
            }
            
            // 填充分类选项
            const categorySelect = document.getElementById('article-category')
            if (categorySelect && result.result.category) {
                categories.length = 0
                // 转换数据格式以匹配新的分类选择器
                const formattedCategories = result.result.category.map(category => ({
                    id: category.value,
                    name: category.name
                }))
                categories.push(...formattedCategories)
                
                // 为隐藏的select添加选项（用于表单提交）
                categorySelect.innerHTML = ''
                result.result.category.forEach(category => {
                    const option = document.createElement('option')
                    option.value = category.value
                    option.textContent = category.name
                    categorySelect.appendChild(option)
                })
                
                // 初始化新的分类选择器
                initCategorySelector()
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
            
            // 使用新的分类选择器设置分类
            if (articleData.categoryId && articleData.categoryId.length > 0) {
                setCategorySelection(articleData.categoryId)
            }
            
            // 更新预览和字符数统计
            if (window.updatePreview) {
                window.updatePreview()
            }
            
            // 更新字符数统计
            const charCountElement = document.getElementById('char-count')
            if (charCountElement && markdownEditor) {
                const count = markdownEditor.value.length
                charCountElement.textContent = count.toLocaleString()
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

// ==================== 分类选择器功能 ====================

// 初始化分类选择器
function initCategorySelector() {
    const searchInput = document.getElementById('category-search')
    const popup = document.getElementById('category-popup')
    const overlay = document.getElementById('category-overlay')
    const selectedContainer = document.getElementById('selected-categories')
    const categorySelector = document.querySelector('.category-selector')
    
    if (!searchInput || !popup || !overlay || !selectedContainer || !categorySelector) {
        return
    }
    
    // 搜索输入事件
    searchInput.addEventListener('input', handleCategorySearch)
    
    // 点击分类选择器模块显示浮层
    categorySelector.addEventListener('click', (e) => {
        // 如果点击的是已选标签的删除按钮，不触发浮层
        if (e.target.closest('.remove-tag')) {
            return
        }
        showCategoryPopup()
    })
    
    // 点击遮罩层关闭浮层
    overlay.addEventListener('click', hideCategoryPopup)
    
    // 点击外部关闭浮层
    document.addEventListener('click', (e) => {
        // 检查点击是否在分类选择器、弹窗内部或者是分类选项
        if (!e.target.closest('.category-selector') && 
            !e.target.closest('#category-popup') && 
            !e.target.closest('.category-option')) {
            hideCategoryPopup()
        }
    })
    
    // ESC键关闭浮层
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            hideCategoryPopup()
        }
    })
    
    // 初始化时渲染所有分类
    categoryConfig.filteredCategories = [...categories]
    renderCategoryOptions()
}

// 处理分类搜索
function handleCategorySearch(e) {
    const searchTerm = e.target.value.toLowerCase().trim()
    
    if (searchTerm === '') {
        categoryConfig.filteredCategories = [...categories]
    } else {
        categoryConfig.filteredCategories = categories.filter(category => 
            category.name.toLowerCase().includes(searchTerm)
        )
    }
    
    renderCategoryOptions()
    
    // 显示或隐藏"无结果"提示
    const noResultsEl = document.getElementById('no-results')
    if (noResultsEl) {
        if (categoryConfig.filteredCategories.length === 0 && searchTerm !== '') {
            noResultsEl.classList.remove('hidden')
        } else {
            noResultsEl.classList.add('hidden')
        }
    }
}

// 显示分类选择浮层
function showCategoryPopup() {
    const popup = document.getElementById('category-popup')
    const overlay = document.getElementById('category-overlay')
    const searchInput = document.getElementById('category-search')
    
    if (popup && overlay) {
        overlay.classList.remove('hidden')
        popup.classList.remove('hidden')
        
        // 聚焦到搜索框
        setTimeout(() => {
            if (searchInput) {
                searchInput.focus()
            }
        }, 100)
    }
}

// 隐藏分类选择浮层
function hideCategoryPopup() {
    const popup = document.getElementById('category-popup')
    const overlay = document.getElementById('category-overlay')
    const searchInput = document.getElementById('category-search')
    
    if (popup && overlay) {
        popup.classList.add('hidden')
        overlay.classList.add('hidden')
    }
    
    // 清空搜索内容并重置过滤结果
    if (searchInput) {
        searchInput.value = ''
        searchInput.blur()
        // 重置过滤结果
        categoryConfig.filteredCategories = [...categories]
        renderCategoryOptions()
    }
}

// 渲染分类选项
function renderCategoryOptions() {
    const optionsContainer = document.getElementById('category-options')
    if (!optionsContainer) return
    
    const filteredCategories = categoryConfig.filteredCategories
    
    optionsContainer.innerHTML = filteredCategories.map(category => {
        const isSelected = categoryConfig.selectedCategories.has(category.id)
        const selectedClass = isSelected ? 'bg-primary text-primary-content' : 'text-base-content'
        const hoverClass = isSelected ? 'hover:bg-primary-focus hover:text-primary-content' : 'hover:bg-base-200 hover:text-base-content'
        
        return `
            <div class="category-option p-2 cursor-pointer rounded transition-colors ${selectedClass} ${hoverClass}" data-category-id="${category.id}">
                <div class="flex items-center justify-between">
                    <span>${category.name}</span>
                    ${isSelected ? '<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd"></path></svg>' : ''}
                </div>
            </div>
        `
    }).join('')
    
    // 添加点击事件
    optionsContainer.querySelectorAll('.category-option').forEach(option => {
        option.addEventListener('click', () => {
            const categoryId = parseInt(option.dataset.categoryId)
            selectCategory(categoryId)
        })
    })
}

// 选择分类
function selectCategory(categoryId) {
    const category = categories.find(c => c.id === categoryId)
    if (!category) return
    
    // 检查是否已选择
    if (categoryConfig.selectedCategories.has(categoryId)) {
        removeCategorySelection(categoryId)
    } else {
        addCategorySelection(category)
    }
    
    // 不再自动关闭浮层，让用户可以继续选择
    // hideCategoryPopup()
}

// 添加分类选择
function addCategorySelection(category) {
    if (categoryConfig.selectedCategories.size >= categoryConfig.maxSelection) {
        showMessage(`最多只能选择${categoryConfig.maxSelection}个分类`, 'error')
        return
    }
    
    categoryConfig.selectedCategories.add(category.id)
    articleData.categoryId = Array.from(categoryConfig.selectedCategories)
    
    updateSelectedCategoriesDisplay()
    updateHiddenSelect()
    renderCategoryOptions()
}

// 移除分类选择
function removeCategorySelection(categoryId) {
    categoryConfig.selectedCategories.delete(categoryId)
    articleData.categoryId = Array.from(categoryConfig.selectedCategories)
    
    updateSelectedCategoriesDisplay()
    updateHiddenSelect()
    renderCategoryOptions()
}

// 更新已选分类的显示
function updateSelectedCategoriesDisplay() {
    const container = document.getElementById('selected-categories')
    const placeholder = document.getElementById('category-placeholder')
    
    if (!container) return
    
    // 清空容器，但保留占位符
    const tags = container.querySelectorAll('.category-tag')
    tags.forEach(tag => tag.remove())
    
    if (categoryConfig.selectedCategories.size === 0) {
        if (placeholder) placeholder.classList.remove('hidden')
        return
    }
    
    if (placeholder) placeholder.classList.add('hidden')
    
    // 添加选中的分类标签
    categoryConfig.selectedCategories.forEach(categoryId => {
        const category = categories.find(c => c.id === categoryId)
        if (!category) return
        
        const tagEl = document.createElement('span')
        tagEl.className = 'category-tag inline-flex items-center gap-1 px-2 py-1 bg-primary text-primary-content text-sm rounded-full'
        tagEl.innerHTML = `
            <span>${category.name}</span>
            <button type="button" class="remove-tag hover:bg-primary-focus rounded-full p-0.5 transition-colors" data-category-id="${categoryId}">
                <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                    <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                </svg>
            </button>
        `
        
        // 添加删除事件
        const removeBtn = tagEl.querySelector('.remove-tag')
        removeBtn.addEventListener('click', (e) => {
            e.stopPropagation()
            const categoryId = parseInt(removeBtn.dataset.categoryId)
            removeCategorySelection(categoryId)
        })
        
        container.appendChild(tagEl)
    })
}

// 更新隐藏的select元素（用于表单提交）
function updateHiddenSelect() {
    const hiddenSelect = document.getElementById('article-category')
    if (!hiddenSelect) return
    
    // 清空现有选项
    hiddenSelect.innerHTML = ''
    
    // 添加选中的分类
    categoryConfig.selectedCategories.forEach(categoryId => {
        const option = document.createElement('option')
        option.value = categoryId
        option.selected = true
        hiddenSelect.appendChild(option)
    })
}

// 设置分类选择器的值（用于编辑模式）
function setCategorySelection(categoryIds) {
    categoryConfig.selectedCategories.clear()
    
    if (Array.isArray(categoryIds)) {
        categoryIds.forEach(id => {
            if (categories.find(c => c.id === id)) {
                categoryConfig.selectedCategories.add(id)
            }
        })
    }
    
    articleData.categoryId = Array.from(categoryConfig.selectedCategories)
    updateSelectedCategoriesDisplay()
    updateHiddenSelect()
    renderCategoryOptions()
}
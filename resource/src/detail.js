import './style.css'
import mermaid from 'mermaid'

// 初始化 mermaid
mermaid.initialize({
    startOnLoad: true,
    theme: 'default',
    securityLevel: 'loose',
    fontFamily: 'inherit'
})

// 渲染 mermaid 图表的函数
function renderMermaidDiagrams() {
    // 查找所有 class 为 language-mermaid 的代码块
    const mermaidBlocks = document.querySelectorAll('code.language-mermaid, pre.language-mermaid code')

    mermaidBlocks.forEach((block, index) => {
        // 创建一个唯一的 ID
        const id = `mermaid-diagram-${index}-${Date.now()}`

        // 创建一个新的 div 元素来替换代码块
        const mermaidDiv = document.createElement('div')
        mermaidDiv.className = 'mermaid'
        mermaidDiv.id = id
        mermaidDiv.textContent = block.textContent

        // 替换原来的代码块
        const parent = block.closest('pre') || block
        parent.parentNode.replaceChild(mermaidDiv, parent)
    })

    // 重新初始化 mermaid 以渲染新添加的图表
    mermaid.init(undefined, '.mermaid')
}

// 页面加载完成后执行
document.addEventListener('DOMContentLoaded', () => {
    renderMermaidDiagrams()
})

// 如果页面内容动态更新，可以调用这个函数重新渲染
window.renderMermaidDiagrams = renderMermaidDiagrams

import './style.css'
import { initMermaidRenderer } from './utils/mermaidRenderer'
import './anchor-links.css'

// 初始化 Mermaid 渲染器
initMermaidRenderer()

// 初始化锚点功能
function initAnchorLinks() {
    // 为所有带有 id 的标题添加锚点链接
    const headings = document.querySelectorAll('.prose h1[id], .prose h2[id], .prose h3[id], .prose h4[id], .prose h5[id], .prose h6[id]')
    
    headings.forEach(heading => {
        // 创建锚点链接元素
        const anchor = document.createElement('a')
        anchor.href = `#${heading.id}`
        anchor.className = 'anchor-link'
        anchor.innerHTML = '#'
        anchor.setAttribute('aria-label', `链接到 ${heading.textContent}`)
        
        // 添加点击事件处理平滑滚动
        anchor.addEventListener('click', (e) => {
            e.preventDefault()
            const target = document.getElementById(heading.id)
            if (target) {
                // 平滑滚动到目标位置
                target.scrollIntoView({ 
                    behavior: 'smooth',
                    block: 'start'
                })
                // 更新 URL hash
                history.pushState(null, null, `#${heading.id}`)
            }
        })
        
        // 将锚点链接添加到标题中
        heading.appendChild(anchor)
    })
}

// 处理页面加载时的 hash 跳转
function handleInitialHash() {
    if (window.location.hash) {
        const target = document.querySelector(window.location.hash)
        if (target) {
            // 延迟滚动，确保页面完全加载
            setTimeout(() => {
                target.scrollIntoView({ 
                    behavior: 'smooth',
                    block: 'start'
                })
            }, 100)
        }
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    initAnchorLinks()
    handleInitialHash()
})

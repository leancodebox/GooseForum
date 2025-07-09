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
                // 计算目标位置，考虑 scroll-margin-top: 4rem
                const targetPosition = target.offsetTop;
                // 平滑滚动到目标位置
                window.scrollTo({
                    top: targetPosition,
                    behavior: 'smooth'
                })
                // 更新 URL hash
                history.pushState(null, null, `#${heading.id}`)
            }
        })
        
        // 将锚点链接添加到标题中
        heading.appendChild(anchor)
    })
}


// 删除评论功能
window.deleteReply = async function(replyId) {
    if (!confirm('确定要删除这条评论吗？')) {
        return;
    }
    
    try {
        const response = await fetch('/api/forum/articles-reply-delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                replyId: replyId
            })
        });
        
        if (response.ok) {
            // 删除成功，移除DOM元素
            const replyElement = document.getElementById(`reply_${replyId}`);
            if (replyElement) {
                replyElement.remove();
            }
            // 可以添加成功提示
            console.log('评论删除成功');
        } else {
            const errorData = await response.json();
            alert('删除失败：' + (errorData.message || '未知错误'));
        }
    } catch (error) {
        console.error('删除评论时发生错误:', error);
        alert('删除失败，请稍后重试');
    }
}

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    initAnchorLinks()
})

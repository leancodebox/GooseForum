import './style.css'
import './utils/notification.ts'
import './anchor-links.css'
import 'cropperjs/dist/cropper.css'

import Alpine from 'alpinejs'
import axios from 'axios'
import Cropper from 'cropperjs'
import { renderMermaidDiagrams } from './utils/mermaidRenderer'
import { processImageFile } from './utils/imageUtils'

// Expose libraries and utils to window
window.Alpine = Alpine
window.axios = axios
window.Cropper = Cropper
window.MermaidUtils = { renderMermaidDiagrams }
window.ImageUtils = { processImageFile }

// Initialize Anchor Links
function initAnchorLinks() {
    const headings = document.querySelectorAll('.prose h1[id], .prose h2[id], .prose h3[id], .prose h4[id], .prose h5[id], .prose h6[id]')
    
    headings.forEach(heading => {
        if (heading.querySelector('.anchor-link')) return; // Avoid duplicates

        const anchor = document.createElement('a')
        anchor.href = `#${heading.id}`
        anchor.className = 'anchor-link'
        anchor.innerHTML = '#'
        anchor.setAttribute('aria-label', `链接到 ${heading.textContent}`)

        anchor.addEventListener('click', (e) => {
            e.preventDefault()
            const target = document.getElementById(heading.id)
            if (target) {
                const targetPosition = target.offsetTop;
                window.scrollTo({
                    top: targetPosition,
                    behavior: 'smooth'
                })
                history.pushState(null, null, `#${heading.id}`)
            }
        })

        heading.appendChild(anchor)
    })
}

// Expose initAnchorLinks if needed, or run on load
window.initAnchorLinks = initAnchorLinks

// Start Alpine
Alpine.start()

// Auto-run certain initializers on DOMContentLoaded
document.addEventListener('DOMContentLoaded', () => {
    initAnchorLinks()
    // We can also try to render mermaid automatically if not handled by Alpine
    // renderMermaidDiagrams() 
})

import type { ContentEnhancer } from '@/runtime/content-enhancements'

const mermaidSelector = 'pre > code.language-mermaid'
let diagramSequence = 0
let mermaidPromise: Promise<typeof import('mermaid').default> | undefined
let renderQueue = Promise.resolve()

function loadMermaid() {
  if (!mermaidPromise) {
    mermaidPromise = import('mermaid').then(({ default: mermaid }) => {
      mermaid.initialize({
        startOnLoad: false,
        securityLevel: 'strict',
        suppressErrorRendering: true,
        fontFamily: 'inherit',
      })
      return mermaid
    })
  }
  return mermaidPromise
}

function renderDiagram(mermaid: typeof import('mermaid').default, id: string, source: string) {
  const result = renderQueue.then(() => mermaid.render(id, source))
  renderQueue = result.then(() => undefined, () => undefined)
  return result
}

async function enhanceMermaid(root: HTMLElement) {
  const blocks = Array.from(root.querySelectorAll<HTMLElement>(mermaidSelector))
    .filter(block => block.parentElement?.dataset.gfEnhancement !== 'mermaid')
  if (!blocks.length) return

  const mermaid = await loadMermaid()
  for (const block of blocks) {
    const sourceElement = block.parentElement
    if (!sourceElement || !sourceElement.isConnected || sourceElement.dataset.gfEnhancement) continue

    sourceElement.dataset.gfEnhancement = 'mermaid'
    try {
      const id = `gf-mermaid-${++diagramSequence}`
      const { svg, bindFunctions } = await renderDiagram(mermaid, id, block.textContent || '')
      if (!sourceElement.isConnected || sourceElement.dataset.gfEnhancement !== 'mermaid') continue

      const diagram = document.createElement('div')
      diagram.className = 'gf-content-diagram gf-content-diagram-mermaid'
      diagram.dataset.gfEnhanced = 'mermaid'
      diagram.innerHTML = svg
      bindFunctions?.(diagram)
      sourceElement.replaceWith(diagram)
    } catch (error) {
      sourceElement.dataset.gfEnhancement = 'mermaid-error'
      console.warn('Unable to render Mermaid diagram; preserving its source code.', error)
    }
  }
}

export const mermaidContentEnhancer: ContentEnhancer = {
  name: 'Mermaid',
  enhance: enhanceMermaid,
}

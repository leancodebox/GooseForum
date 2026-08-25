import type { ObjectDirective } from 'vue'
import { mermaidContentEnhancer } from '@/runtime/content-enhancements/mermaid'

export interface ContentEnhancer {
  name: string
  enhance: (root: HTMLElement) => Promise<void> | void
}

const contentEnhancers: ContentEnhancer[] = [
  mermaidContentEnhancer,
]

export async function enhanceRenderedContent(root: HTMLElement) {
  await Promise.all(contentEnhancers.map(async (enhancer) => {
    try {
      await enhancer.enhance(root)
    } catch (error) {
      console.warn(`Unable to apply the ${enhancer.name} content enhancer.`, error)
    }
  }))
}

export const vContentEnhancements: ObjectDirective<HTMLElement, string | undefined> = {
  mounted(element) {
    void enhanceRenderedContent(element)
  },
  updated(element) {
    void enhanceRenderedContent(element)
  },
}

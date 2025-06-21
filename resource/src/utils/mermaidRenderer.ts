/**
 * Mermaid 图表渲染工具
 * 提供可复用的 Mermaid 图表渲染功能
 */

// Mermaid 实例类型定义
interface MermaidAPI {
  initialize(config: any): void;
  run(options: { querySelector: string }): Promise<void>;
}

/**
 * 检查页面是否存在需要渲染的 Mermaid 代码块
 * @returns {boolean} 是否存在 Mermaid 代码块
 */
export function hasMermaidBlocks(): boolean {
  const mermaidBlocks = document.querySelectorAll('code.language-mermaid, pre.language-mermaid code');
  return mermaidBlocks.length > 0;
}

/**
 * 动态加载并初始化 Mermaid
 * @returns {Promise<MermaidAPI | null>} Mermaid 实例或 null
 */
export async function loadMermaid(): Promise<MermaidAPI | null> {
  try {
    const { default: mermaid } = await import('mermaid');
    
    // 初始化 mermaid
    mermaid.initialize({
      startOnLoad: false, // 设置为 false，手动控制渲染
      theme: 'default',
      securityLevel: 'loose',
      fontFamily: 'inherit'
    });
    
    return mermaid;
  } catch (error) {
    console.error('Failed to load Mermaid:', error);
    return null;
  }
}

/**
 * 渲染 Mermaid 图表
 * @param {string} containerSelector 可选的容器选择器，默认为整个文档
 * @returns {Promise<void>}
 */
export async function renderMermaidDiagrams(containerSelector?: string): Promise<void> {
  // 首先检查是否存在需要渲染的 Mermaid 代码块
  if (!hasMermaidBlocks()) {
    console.log('No Mermaid blocks found, skipping Mermaid loading');
    return;
  }

  console.log('Mermaid blocks detected, loading Mermaid library...');
  
  // 动态加载 Mermaid
  const mermaid = await loadMermaid();
  if (!mermaid) {
    console.error('Failed to load Mermaid library');
    return;
  }

  // 确定搜索范围
  const searchRoot = containerSelector ? document.querySelector(containerSelector) : document;
  if (!searchRoot) {
    console.error(`Container not found: ${containerSelector}`);
    return;
  }

  // 查找所有 class 为 language-mermaid 的代码块
  const mermaidBlocks = searchRoot.querySelectorAll('code.language-mermaid, pre.language-mermaid code');

  mermaidBlocks.forEach((block, index) => {
    // 创建一个唯一的 ID
    const id = `mermaid-diagram-${index}-${Date.now()}`;
    
    // 创建一个新的 div 元素来替换代码块
    const mermaidDiv = document.createElement('div');
    mermaidDiv.className = 'mermaid';
    mermaidDiv.id = id;
    mermaidDiv.textContent = block.textContent;

    // 替换原来的代码块
    const parent = block.closest('pre') || block;
    if (parent.parentNode) {
      parent.parentNode.replaceChild(mermaidDiv, parent);
    }
  });

  // 使用现代的 mermaid.run API 渲染图表
  try {
    await mermaid.run({
      querySelector: containerSelector ? `${containerSelector} .mermaid` : '.mermaid'
    });
    console.log('Mermaid diagrams rendered successfully');
  } catch (error) {
    console.error('Failed to render Mermaid diagrams:', error);
  }
}

/**
 * 初始化 Mermaid 渲染器
 * 在 DOM 加载完成后自动渲染 Mermaid 图表
 * @param {string} containerSelector 可选的容器选择器
 */
export function initMermaidRenderer(containerSelector?: string): void {
  // 页面加载完成后执行
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', async () => {
      await renderMermaidDiagrams(containerSelector);
    });
  } else {
    // 如果 DOM 已经加载完成，直接执行
    renderMermaidDiagrams(containerSelector);
  }

  // 将渲染函数暴露到全局，便于动态内容更新时调用
  (window as any).renderMermaidDiagrams = renderMermaidDiagrams;
}

/**
 * 默认导出：简化的初始化函数
 */
export default {
  init: initMermaidRenderer,
  render: renderMermaidDiagrams,
  hasBlocks: hasMermaidBlocks,
  loadMermaid
};
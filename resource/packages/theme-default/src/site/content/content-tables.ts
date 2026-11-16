// Keep native table semantics; only the surrounding region scrolls.
export function enhanceContentTables(root: HTMLElement) {
  const cleanups = Array.from(root.querySelectorAll<HTMLTableElement>("table")).map(table => {
    const shell = document.createElement("div");
    shell.className = "gf-table-shell";
    const viewport = document.createElement("div");
    viewport.className = "gf-table-scroll";
    const headings = Array.from(table.querySelectorAll("thead th"), cell => cell.textContent?.trim()).filter(Boolean);
    if (headings.length) {
      viewport.setAttribute("role", "region");
      viewport.setAttribute("aria-label", headings.join(" · "));
    }
    table.before(shell);
    shell.append(viewport);
    viewport.append(table);

    // Only uniformly short columns stay compact. Never infer a column's role from its position.
    const rows = Array.from(table.rows);
    // Cell indexes do not represent visual columns when any row contains spans.
    const hasMergedCells = rows.some(row => Array.from(row.cells).some(cell => cell.colSpan !== 1 || cell.rowSpan !== 1));
    const compactCells = new Set<HTMLTableCellElement>();
    for (let column = 0; !hasMergedCells && column < (rows[0]?.cells.length ?? 0); column++) {
      const cells = rows.map(row => row.cells[column]);
      if (cells.some(cell => !cell || cell.colSpan !== 1 || cell.rowSpan !== 1)) continue;
      if (cells.every(cell => (cell.textContent?.trim().length ?? 0) <= 6 && !cell.querySelector("img, pre, br"))) {
        cells.forEach(cell => { cell.classList.add("gf-table-compact"); compactCells.add(cell); });
      }
    }
    const update = () => {
      const overflow = viewport.scrollWidth > viewport.clientWidth + 1;
      viewport.tabIndex = overflow ? 0 : -1;
      shell.dataset.scrollLeft = String(overflow && viewport.scrollLeft > 1);
      shell.dataset.scrollRight = String(overflow && viewport.scrollLeft + viewport.clientWidth < viewport.scrollWidth - 1);
    };
    viewport.addEventListener("scroll", update, { passive: true });
    const observer = typeof ResizeObserver === "undefined" ? undefined : new ResizeObserver(update);
    observer?.observe(viewport);
    observer?.observe(table);
    update();
    return () => {
      observer?.disconnect();
      viewport.removeEventListener("scroll", update);
      compactCells.forEach(cell => cell.classList.remove("gf-table-compact"));
      // React may already have replaced the HTML when this effect is cleaned up.
      if (shell.parentNode) shell.replaceWith(table);
    };
  });
  return () => cleanups.forEach(cleanup => cleanup());
}

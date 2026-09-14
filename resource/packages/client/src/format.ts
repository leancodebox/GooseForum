export function formatCompactNumber(value: number): string {
  if (value >= 1_000_000) return `${trimCompactNumber(value / 1_000_000)}m`
  if (value >= 1_000) return `${trimCompactNumber(value / 1_000)}k`
  return String(value)
}

function trimCompactNumber(value: number) {
  return value.toFixed(value >= 10 ? 0 : 1).replace(/\.0$/, '')
}

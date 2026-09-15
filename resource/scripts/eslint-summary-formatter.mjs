export default function format(results) {
  const counts = new Map()
  const errors = []
  let warningCount = 0

  for (const result of results) {
    for (const message of result.messages) {
      const rule = message.ruleId || 'configuration'
      counts.set(rule, (counts.get(rule) || 0) + 1)
      if (message.severity === 2) {
        errors.push(`${result.filePath}:${message.line}:${message.column} ${rule} ${message.message}`)
      } else {
        warningCount += 1
      }
    }
  }

  const summary = [...counts.entries()]
    .sort((left, right) => right[1] - left[1] || left[0].localeCompare(right[0]))
    .map(([rule, count]) => `  ${rule}: ${count}`)

  return [
    ...errors,
    `ESLint: ${errors.length} error(s), ${warningCount} warning(s)`,
    ...summary,
  ].join('\n')
}

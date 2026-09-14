import type { ReactNode } from 'react'

export function PageHeader({
  title,
  description,
  badge,
  actions,
}: {
  title: string
  description?: string
  badge?: ReactNode
  actions?: ReactNode
}) {
  return (
    <header className="flex flex-col gap-3 border-b px-4 py-4 lg:flex-row lg:items-start lg:justify-between lg:px-0 lg:py-3">
      <div className="min-w-0">
        <div className="flex min-w-0 flex-wrap items-center gap-2">
          <h1 className="min-w-0 truncate text-xl font-bold lg:text-2xl">{title}</h1>
          {badge}
        </div>
        {description ? <p className="mt-1 max-w-2xl text-sm leading-6 text-muted-foreground">{description}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 flex-wrap items-center gap-2">{actions}</div> : null}
    </header>
  )
}

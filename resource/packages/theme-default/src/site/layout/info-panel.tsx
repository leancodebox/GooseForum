import type { ReactNode } from 'react'

export function InfoPanel({ title, children, action }: { title: string, children: ReactNode, action?: ReactNode }) {
  return (
    <section className="site-panel rounded-xl border bg-background p-4">
      <h2 className="text-sm font-semibold">{title}</h2>
      <div className="mt-2 text-sm leading-6 text-muted-foreground">{children}</div>
      {action ? <div className="mt-4">{action}</div> : null}
    </section>
  )
}

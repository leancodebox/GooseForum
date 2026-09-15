import type { LayoutPayload } from '@gooseforum/client'
import { GooseLink } from '@gooseforum/runtime'

export function AuthBrand({ layout }: { layout: LayoutPayload }) {
  const { site } = layout
  let content
  if (site.brandType === 'image' && site.brandImage) {
    content = <img src={site.brandImage} alt={site.name} className="h-8 w-auto object-contain" />
  } else if (site.brandType === 'text') {
    content = site.brandText || site.name
  } else {
    content = <>Goose<span className="text-foreground">Forum</span></>
  }

  return (
    <GooseLink href="/" className="mb-3 w-fit text-2xl font-semibold tracking-tight text-primary">
      {content}
    </GooseLink>
  )
}

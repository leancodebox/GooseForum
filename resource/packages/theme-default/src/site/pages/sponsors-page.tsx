import { ExternalLinkIcon, HeartHandshakeIcon, MailIcon, ShieldCheckIcon } from 'lucide-react'
import type { SponsorSectionPayload, SponsorsPageProps } from '@gooseforum/client'
import { useTranslation } from 'react-i18next'
import { Badge } from '@gooseforum/ui/components/badge'
import { Button } from '@gooseforum/ui/components/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@gooseforum/ui/components/empty'
import { cn } from '@gooseforum/ui/lib/utils'
import { InfoPanel } from '../layout/info-panel'
import { PageHeader } from '../layout/page-header'

export function SponsorsPageView({ page }: { page: SponsorsPageProps }) {
  const { t } = useTranslation('sponsors')
  return (
    <div className="pb-12">
      <PageHeader compact divided={false} title={page.content.title} description={page.content.description} badge={<Badge variant="secondary">{page.totalCount}</Badge>} />
      <div className="grid gap-0 lg:gap-5 xl:grid-cols-[minmax(0,1fr)_260px]">
        <div className="flex min-w-0 flex-col gap-0 lg:gap-6">
          {page.sections.map((section) => (
            <section key={section.key} className="flex flex-col gap-0 lg:gap-2.5">
              <div className="flex items-center justify-between gap-3 px-4 py-3 lg:px-0 lg:py-0">
                <h2>
                  <Badge
                    variant="secondary"
                    className={cn(
                      section.tone === 'diamond' && 'bg-primary/10 text-primary',
                      section.tone === 'gold' && 'bg-warning/10 text-warning',
                    )}
                  >
                    {section.label}
                  </Badge>
                </h2>
                <Badge variant="secondary">{section.sponsors.length}</Badge>
              </div>
              <div className={sectionGrid(section)}>
                {section.sponsors.map((sponsor) => {
                  const content = (
                    <>
                      <img src={sponsor.avatarUrl} alt={sponsor.name} className={cn('shrink-0 rounded-lg border object-cover', avatarSize(section))} loading="lazy" />
                      <div className="min-w-0 flex-1">
                        <div className="flex min-w-0 items-center gap-1.5">
                          <h3 className="truncate text-[13px] font-semibold group-hover:text-primary lg:text-sm">{sponsor.name}</h3>
                          {sponsor.link ? <ExternalLinkIcon width={14} height={14} className="shrink-0 text-muted-foreground group-hover:text-primary" aria-hidden="true" /> : null}
                        </div>
                        <p className="mt-0.5 line-clamp-2 text-[11px] leading-4 text-muted-foreground lg:mt-1 lg:text-xs lg:leading-5">
                          {sponsor.message || t('defaultMessage')}
                        </p>
                      </div>
                    </>
                  )
                  const className = cn('group flex items-start gap-2.5 border-b bg-background lg:rounded-xl lg:border transition-colors hover:border-primary/30 hover:bg-primary/5', cardPadding(section))
                  return sponsor.link
                    ? <a key={`${section.key}-${sponsor.name}`} href={sponsor.link} target="_blank" rel="noopener noreferrer" className={className}>{content}</a>
                    : <div key={`${section.key}-${sponsor.name}`} className={className}>{content}</div>
                })}
              </div>
            </section>
          ))}
          {!page.sections.length
            ? (
                <Empty className="site-panel min-h-56 border bg-background">
                  <EmptyHeader>
                    <EmptyMedia variant="icon"><HeartHandshakeIcon /></EmptyMedia>
                    <EmptyTitle>{t('emptyTitle')}</EmptyTitle>
                    <EmptyDescription>{t('emptyDescription')}</EmptyDescription>
                  </EmptyHeader>
                </Empty>
              )
            : null}
        </div>

        <aside className="flex flex-col gap-0 lg:gap-3">
          <InfoPanel
            title={page.contact.title}
            action={<Button asChild><a href={page.contact.buttonLink}><MailIcon data-icon="inline-start" />{page.contact.buttonText}</a></Button>}
          >
            {page.contact.description}
          </InfoPanel>
          {page.rules.length
            ? (
                <InfoPanel title={t('rulesTitle')}>
                  <ul className="flex flex-col gap-2 text-foreground/75">
                    {page.rules.map((rule) => (
                      <li key={rule.content} className="flex gap-2">
                        <ShieldCheckIcon width={16} height={16} className="mt-0.5 shrink-0 text-success" aria-hidden="true" />
                        <span>{rule.content}</span>
                      </li>
                    ))}
                  </ul>
                </InfoPanel>
              )
            : null}
        </aside>
      </div>
    </div>
  )
}

function sectionGrid(section: SponsorSectionPayload) {
  return cn(
    'grid grid-cols-1 gap-0 lg:gap-3',
    section.tone === 'diamond' && 'lg:grid-cols-2',
    section.tone === 'gold' && 'lg:grid-cols-2 xl:grid-cols-3',
    section.tone !== 'diamond' && section.tone !== 'gold' && 'lg:grid-cols-4 2xl:grid-cols-5',
  )
}

function cardPadding(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'p-3.5'
  if (section.tone === 'gold') return 'p-3'
  return 'p-2.5'
}

function avatarSize(section: SponsorSectionPayload) {
  if (section.tone === 'diamond') return 'size-11'
  if (section.tone === 'gold') return 'size-10'
  return 'size-8'
}

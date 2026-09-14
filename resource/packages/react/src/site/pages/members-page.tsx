import { CalendarDaysIcon, ChevronLeftIcon, ChevronRightIcon, UsersRoundIcon } from 'lucide-react'
import { formatCompactNumber, type MembersPageProps } from '@gooseforum/client'
import { useTranslation } from 'react-i18next'
import { Avatar, AvatarFallback, AvatarImage } from '../../components/ui/avatar'
import { Button } from '../../components/ui/button'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '../../components/ui/empty'
import { GooseLink } from '../../runtime'
import { PageHeader } from '../layout/page-header'

export function MembersPageView({ page }: { page: MembersPageProps }) {
  const { t } = useTranslation('members')
  return (
    <div className="pb-12">
      <PageHeader title={t('title')} description={t('subtitle')} />
      {page.members.length
        ? (
            <section className="grid lg:grid-cols-2 lg:gap-3 lg:pt-3 xl:grid-cols-3">
              {page.members.map((member) => {
                const displayName = member.nickname || member.username
                return (
                  <GooseLink
                    key={member.id}
                    href={member.url}
                    className="group flex min-w-0 flex-col overflow-hidden border-b bg-background transition-colors hover:border-primary/30 hover:bg-muted lg:rounded-xl lg:border"
                  >
                    <div className="flex min-w-0 items-center gap-2.5 px-3.5 pt-3.5">
                      <Avatar size="lg" className="ring-1 ring-border transition group-hover:ring-primary/30">
                        <AvatarImage src={member.avatarUrl} alt={displayName} />
                        <AvatarFallback>{displayName.slice(0, 2).toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="min-w-0 flex-1">
                        <div className="flex min-w-0 items-baseline gap-1.5">
                          <h2 className="truncate text-sm font-bold transition-colors group-hover:text-primary">{displayName}</h2>
                          <span className="truncate text-[11px] text-muted-foreground/80">@{member.username}</span>
                        </div>
                        <span className="mt-0.5 inline-flex items-center gap-1 text-[11px] text-muted-foreground">
                          <CalendarDaysIcon width={12} height={12} aria-hidden="true" />
                          {t('joinedAt', { date: member.joinedAt })}
                        </span>
                      </div>
                      <ChevronRightIcon width={16} height={16} className="shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5 group-hover:text-primary" aria-hidden="true" />
                    </div>
                    <p className="mx-3.5 mb-3 mt-2 line-clamp-2 min-h-10 text-xs leading-5 text-muted-foreground">{member.bio || t('noBio')}</p>
                    <div className="mt-auto grid grid-cols-3 divide-x border-t bg-muted/50 py-2">
                      <MemberStat value={member.prestige} label={t('prestige')} />
                      <MemberStat value={member.topicCount} label={t('topics')} />
                      <MemberStat value={member.replyCount} label={t('replies')} />
                    </div>
                  </GooseLink>
                )
              })}
            </section>
          )
        : (
            <Empty className="min-h-56 border bg-background">
              <EmptyHeader>
                <EmptyMedia variant="icon"><UsersRoundIcon /></EmptyMedia>
                <EmptyTitle>{t('emptyTitle')}</EmptyTitle>
                <EmptyDescription>{t('emptyDescription')}</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}

      {page.previousUrl || page.pagination.hasNext
        ? (
            <nav className="mt-4 flex items-center justify-between gap-3 border-t px-4 pt-4 lg:px-0" aria-label={t('pagination')}>
              {page.previousUrl
                ? <Button asChild variant="outline" size="sm"><GooseLink href={page.previousUrl} rel="prev"><ChevronLeftIcon data-icon="inline-start" />{t('previous')}</GooseLink></Button>
                : <span />}
              {page.pagination.hasNext
                ? <Button asChild variant="outline" size="sm"><GooseLink href={page.pagination.nextUrl} rel="next">{t('next')}<ChevronRightIcon data-icon="inline-end" /></GooseLink></Button>
                : <span />}
            </nav>
          )
        : null}
    </div>
  )
}

function MemberStat({ value, label }: { value: number, label: string }) {
  return (
    <span className="flex min-w-0 flex-col items-center justify-center gap-1 px-1 text-xs text-muted-foreground">
      <strong className="max-w-full truncate text-sm font-semibold text-foreground/75">{formatCompactNumber(value)}</strong>
      <span className="text-center text-[11px]">{label}</span>
    </span>
  )
}

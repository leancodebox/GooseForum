'use client'

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format, subDays } from 'date-fns'
import { DateRange } from 'react-day-picker'
import { ContentLayout } from '@/components/layout/content-layout'
import { Users, FileText, MessageSquare, Link as LinkIcon } from 'lucide-react'
import { TrafficOverview } from './components/traffic-overview'
import { DateRangePicker } from '@/components/date-range-picker'
import { ProjectVersion } from './components/project-version'
import { getSiteStatistics, getTrafficOverview } from '@/api'

interface Release {
  id: number
  tag_name: string
  published_at: string
  body: string
  html_url: string
  prerelease: boolean
  draft: boolean
}

export function DashboardStats() {
  const [dateRange, setDateRange] = useState<DateRange | undefined>({
    from: subDays(new Date(), 7),
    to: new Date(),
  })

  const { data: statsData, isLoading: statsLoading } = useQuery({
    queryKey: ['siteStatistics'],
    queryFn: () => getSiteStatistics(),
  })

  const { data: trafficData, isLoading: trafficLoading } = useQuery({
    queryKey: ['trafficOverview', dateRange?.from, dateRange?.to],
    queryFn: () => getTrafficOverview({
      startDate: dateRange?.from ? format(dateRange.from, 'yyyy-MM-dd') : undefined,
      endDate: dateRange?.to ? format(dateRange.to, 'yyyy-MM-dd') : undefined,
    }),
    enabled: !!dateRange?.from && !!dateRange?.to,
  })

  const [releases, setReleases] = useState<Release[]>([])
  const [releasesLoading, setReleasesLoading] = useState(false)
  const [releasesError, setReleasesError] = useState<string | null>(null)

  useEffect(() => {
    const fetchReleases = async () => {
      setReleasesLoading(true)
      try {
        const response = await fetch('https://api.github.com/repos/leancodebox/GooseForum/releases')
        if (response.ok) {
          const data = await response.json()
          setReleases(data)
        } else {
          setReleasesError('无法获取版本信息')
        }
      } catch (error) {
        console.error('Failed to fetch releases:', error)
        setReleasesError('网络错误，无法连接到 GitHub')
      } finally {
        setReleasesLoading(false)
      }
    }
    fetchReleases()
  }, [])

  const stats = statsData?.result

  return (
    <ContentLayout
      title='站点统计'
      description='查看论坛的实时运行数据和活跃度指标。'
      topNav={topNav}
      headerActions={
        <div className='flex items-center gap-6 mt-4 md:mt-0'>
          <div className='flex flex-col items-end justify-center'>
            <div className='flex items-center gap-1.5 text-muted-foreground mb-1'>
              <Users className='h-3.5 w-3.5' />
              <span className='text-xs font-medium'>总用户数</span>
            </div>
            <div className='flex items-baseline gap-1.5'>
              <span className='text-xl font-bold leading-none'>{statsLoading ? '...' : stats?.userCount.toLocaleString()}</span>
              <span className='text-[10px] text-muted-foreground'>+{stats?.userMonthCount || 0}</span>
            </div>
          </div>
          
          <div className='h-8 w-px bg-border' />

          <div className='flex flex-col items-end justify-center'>
            <div className='flex items-center gap-1.5 text-muted-foreground mb-1'>
              <FileText className='h-3.5 w-3.5' />
              <span className='text-xs font-medium'>总文章数</span>
            </div>
            <div className='flex items-baseline gap-1.5'>
              <span className='text-xl font-bold leading-none'>{statsLoading ? '...' : stats?.articleCount.toLocaleString()}</span>
              <span className='text-[10px] text-muted-foreground'>+{stats?.articleMonthCount || 0}</span>
            </div>
          </div>

          <div className='h-8 w-px bg-border hidden sm:block' />

          <div className='hidden sm:flex flex-col items-end justify-center'>
            <div className='flex items-center gap-1.5 text-muted-foreground mb-1'>
              <MessageSquare className='h-3.5 w-3.5' />
              <span className='text-xs font-medium'>总回复数</span>
            </div>
            <div className='flex items-baseline gap-1.5'>
              <span className='text-xl font-bold leading-none'>{statsLoading ? '...' : stats?.reply.toLocaleString()}</span>
            </div>
          </div>

          <div className='h-8 w-px bg-border hidden md:block' />

          <div className='hidden md:flex flex-col items-end justify-center'>
            <div className='flex items-center gap-1.5 text-muted-foreground mb-1'>
              <LinkIcon className='h-3.5 w-3.5' />
              <span className='text-xs font-medium'>友情链接</span>
            </div>
            <div className='flex items-baseline gap-1.5'>
              <span className='text-xl font-bold leading-none'>{statsLoading ? '...' : stats?.linksCount}</span>
            </div>
          </div>
        </div>
      }
    >
      <div className='mt-2 grid grid-cols-1 gap-8 lg:grid-cols-12'>
          <div className='lg:col-span-8'>
            <TrafficOverview 
              data={trafficData?.result || []} 
              loading={trafficLoading} 
              headerAction={<DateRangePicker date={dateRange} setDate={setDateRange} />}
            />
          </div>
          <div className='lg:col-span-4'>
            <ProjectVersion
              releases={releases}
              loading={releasesLoading}
              error={releasesError}
            />
          </div>
        </div>
      </ContentLayout>
    )
  }

const topNav = [
  {
    title: '站点概览',
    href: '/dashboard-stats',
    isActive: true,
    disabled: false,
  },
]

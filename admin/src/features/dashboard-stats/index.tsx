'use client'

import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { format, subDays } from 'date-fns'
import { DateRange } from 'react-day-picker'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
    >
      <div className='grid gap-4 sm:grid-cols-2 lg:grid-cols-4'>
          <Card>
            <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
              <CardTitle className='text-sm font-medium'>总用户数</CardTitle>
              <Users className='h-4 w-4 text-muted-foreground' />
            </CardHeader>
            <CardContent>
              <div className='text-2xl font-bold'>{statsLoading ? '...' : stats?.userCount.toLocaleString()}</div>
              <p className='text-xs text-muted-foreground'>
                本月新增: {statsLoading ? '...' : stats?.userMonthCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
              <CardTitle className='text-sm font-medium'>总文章数</CardTitle>
              <FileText className='h-4 w-4 text-muted-foreground' />
            </CardHeader>
            <CardContent>
              <div className='text-2xl font-bold'>{statsLoading ? '...' : stats?.articleCount.toLocaleString()}</div>
              <p className='text-xs text-muted-foreground'>
                本月新增: {statsLoading ? '...' : stats?.articleMonthCount}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
              <CardTitle className='text-sm font-medium'>总回复数</CardTitle>
              <MessageSquare className='h-4 w-4 text-muted-foreground' />
            </CardHeader>
            <CardContent>
              <div className='text-2xl font-bold'>{statsLoading ? '...' : stats?.reply.toLocaleString()}</div>
              <p className='text-xs text-muted-foreground'>
                社区活跃度指标
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
              <CardTitle className='text-sm font-medium'>友情链接</CardTitle>
              <LinkIcon className='h-4 w-4 text-muted-foreground' />
            </CardHeader>
            <CardContent>
              <div className='text-2xl font-bold'>{statsLoading ? '...' : stats?.linksCount}</div>
              <p className='text-xs text-muted-foreground'>
                合作站点数量
              </p>
            </CardContent>
          </Card>
        </div>

        <div className='mt-8 grid grid-cols-1 gap-8 lg:grid-cols-12'>
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

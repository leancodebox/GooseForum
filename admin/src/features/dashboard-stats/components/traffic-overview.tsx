import { ReactNode } from 'react'
import { Area, AreaChart, ResponsiveContainer, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

interface TrafficData {
  date: string
  regCount: number
  articleCount: number
  replyCount: number
}

interface TrafficOverviewProps {
  data: TrafficData[]
  loading?: boolean
  headerAction?: ReactNode
}

export function TrafficOverview({ data, loading, headerAction }: TrafficOverviewProps) {
  return (
    <Card className='h-full'>
      <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-7'>
        <div className='space-y-1'>
          <CardTitle>流量概览</CardTitle>
          <CardDescription>展示注册用户、文章和回复的增长趋势</CardDescription>
        </div>
        {headerAction}
      </CardHeader>
      <CardContent className='px-2 sm:px-6'>
        <div className='h-[350px] w-full' style={{ minWidth: 0 }}>
          {loading ? (
            <div className='flex h-full w-full items-center justify-center text-muted-foreground'>
              加载中...
            </div>
          ) : (
            <ResponsiveContainer width='100%' height={350}>
              <AreaChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
                <defs>
                          <linearGradient id='colorReg' x1='0' y1='0' x2='0' y2='1'>
                            <stop offset='5%' stopColor='#2563eb' stopOpacity={0.3} />
                            <stop offset='95%' stopColor='#2563eb' stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id='colorArticle' x1='0' y1='0' x2='0' y2='1'>
                            <stop offset='5%' stopColor='#10b981' stopOpacity={0.3} />
                            <stop offset='95%' stopColor='#10b981' stopOpacity={0} />
                          </linearGradient>
                          <linearGradient id='colorReply' x1='0' y1='0' x2='0' y2='1'>
                            <stop offset='5%' stopColor='#f59e0b' stopOpacity={0.3} />
                            <stop offset='95%' stopColor='#f59e0b' stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray='3 3' vertical={false} stroke='hsl(var(--muted-foreground))' opacity={0.1} />
                        <XAxis
                          dataKey='date'
                          stroke='hsl(var(--muted-foreground))'
                          fontSize={12}
                          tickLine={false}
                          axisLine={false}
                          tickFormatter={(value) => value.split('-').slice(1).join('/')}
                        />
                        <YAxis
                          stroke='hsl(var(--muted-foreground))'
                          fontSize={12}
                          tickLine={false}
                          axisLine={false}
                        />
                        <Tooltip
                          content={({ active, payload, label }) => {
                            if (active && payload && payload.length) {
                              return (
                                <div className='rounded-lg border bg-background p-2 shadow-sm'>
                                  <div className='mb-2 border-b pb-1'>
                                    <span className='text-[0.70rem] uppercase text-muted-foreground'>
                                      日期：
                                    </span>
                                    <span className='font-bold text-foreground ml-1'>
                                      {label}
                                    </span>
                                  </div>
                                  <div className='flex flex-col gap-1'>
                                    {payload.map((item) => (
                                      <div key={item.name} className='flex items-center justify-between gap-4'>
                                        <div className='flex items-center gap-1.5'>
                                          <div 
                                            className='h-2 w-2 rounded-full' 
                                            style={{ backgroundColor: item.color }}
                                          />
                                          <span className='text-[0.70rem] text-muted-foreground'>
                                            {item.name === 'regCount' ? '注册用户' : item.name === 'articleCount' ? '新增文章' : '新增回复'}
                                          </span>
                                        </div>
                                        <span className='font-bold text-sm' style={{ color: item.color }}>
                                          {item.value}
                                        </span>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              )
                            }
                            return null
                          }}
                        />
                        <Area
                          name='regCount'
                          type='monotone'
                          dataKey='regCount'
                          stroke='#2563eb'
                          strokeWidth={2}
                          fill='url(#colorReg)'
                        />
                        <Area
                          name='articleCount'
                          type='monotone'
                          dataKey='articleCount'
                          stroke='#10b981'
                          strokeWidth={2}
                          fill='url(#colorArticle)'
                        />
                        <Area
                          name='replyCount'
                          type='monotone'
                          dataKey='replyCount'
                          stroke='#f59e0b'
                          strokeWidth={2}
                          fill='url(#colorReply)'
                        />
              </AreaChart>
            </ResponsiveContainer>
          )}
        </div>
      </CardContent>
    </Card>
  )
}

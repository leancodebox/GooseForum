import { useEffect, useState } from 'react'
import axios from 'axios'
import { toast } from 'sonner'
import {
  Search as SearchIcon,
  RefreshCcw,
  Eye,
  Copy,
  MoreVertical,
  ChevronLeft,
  ChevronRight,
  Loader2,
  Ticket as TicketIcon,
} from 'lucide-react'
import { ContentLayout } from '@/components/layout/content-layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Ticket,
  TICKET_TYPES,
  TICKET_STATUS,
  TICKET_TYPE_VARIANTS,
  TICKET_STATUS_VARIANTS,
} from './data/schema'
import { TicketDetailDialog } from './components/ticket-detail-dialog'

export default function TicketManagement() {
  const [tickets, setTickets] = useState<Ticket[]>([])
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize] = useState(10)
  const [searchParams, setSearchParams] = useState({
    title: '',
    type: '',
    status: '',
  })
  const [selectedTicket, setSelectedTicket] = useState<Ticket | null>(null)

  const fetchTickets = async () => {
    setLoading(true)
    try {
      const response = await axios.post('/api/admin/apply-sheet-list', {
        page,
        pageSize,
      })
      if (response.data.code === 0) {
        setTickets(response.data.result.list)
        setTotal(response.data.result.total)
      } else {
        toast.error(response.data.msg || '加载工单列表失败')
      }
    } catch (error) {
      toast.error('加载工单列表出错')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchTickets()
  }, [page])

  const handleReset = () => {
    setSearchParams({ title: '', type: '', status: '' })
    if (page === 1) {
      fetchTickets()
    } else {
      setPage(1)
    }
  }

  const handleCopy = (ticket: Ticket) => {
    const content = `工单 #${ticket.id}: ${ticket.title}\n\n用户信息: ${ticket.applyUserInfo}\n类型: ${TICKET_TYPES[ticket.type as keyof typeof TICKET_TYPES]}\n状态: ${TICKET_STATUS[ticket.status as keyof typeof TICKET_STATUS]}\n\n内容:\n${ticket.content}`
    navigator.clipboard.writeText(content)
    toast.success('内容已复制到剪贴板')
  }

  const filteredTickets = tickets.filter((ticket) => {
    const matchTitle = ticket.title
      .toLowerCase()
      .includes(searchParams.title.toLowerCase())
    const matchType = searchParams.type
      ? ticket.type.toString() === searchParams.type
      : true
    const matchStatus = searchParams.status
      ? ticket.status.toString() === searchParams.status
      : true
    return matchTitle && matchType && matchStatus
  })

  const totalPages = Math.ceil(total / pageSize)

  return (
    <>
      <ContentLayout title='工单管理' description='查看和管理用户提交的工单。'>
      <div className='grid grid-cols-1 gap-4 md:grid-cols-4'>
          <div className='relative'>
            <SearchIcon className='absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground' />
            <Input
              placeholder='搜索工单标题'
              className='pl-8'
              value={searchParams.title}
              onChange={(e) =>
                setSearchParams({ ...searchParams, title: e.target.value })
              }
            />
          </div>
          <Select
            value={searchParams.type}
            onValueChange={(val) => setSearchParams({ ...searchParams, type: val })}
          >
            <SelectTrigger>
              <SelectValue placeholder='全部类型' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>全部类型</SelectItem>
              {Object.entries(TICKET_TYPES).map(([id, label]) => (
                <SelectItem key={id} value={id}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select
            value={searchParams.status}
            onValueChange={(val) => setSearchParams({ ...searchParams, status: val })}
          >
            <SelectTrigger>
              <SelectValue placeholder='全部状态' />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value='all'>全部状态</SelectItem>
              {Object.entries(TICKET_STATUS).map(([id, label]) => (
                <SelectItem key={id} value={id}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button variant='outline' onClick={handleReset}>
            <RefreshCcw className='mr-2 h-4 w-4' />
            重置
          </Button>
        </div>

        <div className='rounded-md border bg-card'>
          {loading ? (
            <div className='flex h-64 items-center justify-center'>
              <Loader2 className='h-8 w-8 animate-spin text-primary' />
            </div>
          ) : filteredTickets.length === 0 ? (
            <div className='flex h-64 flex-col items-center justify-center gap-2'>
              <TicketIcon className='h-12 w-12 text-muted-foreground opacity-20' />
              <p className='text-muted-foreground'>暂无工单数据</p>
            </div>
          ) : (
            <ul className='divide-y'>
              {filteredTickets.map((ticket) => (
                <li
                  key={ticket.id}
                  className='flex items-center justify-between p-4 transition-colors hover:bg-muted/50'
                >
                  <div className='flex min-w-0 flex-1 flex-col gap-1'>
                    <div className='flex items-center gap-2'>
                      <span className='text-sm font-medium text-muted-foreground'>
                        #{ticket.id}
                      </span>
                      <h4 className='truncate font-semibold'>{ticket.title}</h4>
                      <Badge
                        variant={
                          TICKET_TYPE_VARIANTS[
                            ticket.type as keyof typeof TICKET_TYPE_VARIANTS
                          ] as any
                        }
                        className='text-[10px]'
                      >
                        {TICKET_TYPES[ticket.type as keyof typeof TICKET_TYPES]}
                      </Badge>
                      <Badge
                        variant={
                          TICKET_STATUS_VARIANTS[
                            ticket.status as keyof typeof TICKET_STATUS_VARIANTS
                          ] as any
                        }
                        className='text-[10px]'
                      >
                        {TICKET_STATUS[ticket.status as keyof typeof TICKET_STATUS]}
                      </Badge>
                    </div>
                    <div className='flex items-center gap-2 text-xs text-muted-foreground'>
                      <span>{ticket.applyUserInfo}</span>
                      <span>•</span>
                      <span>{new Date(ticket.createTime).toLocaleString()}</span>
                      {ticket.content && (
                        <>
                          <span>•</span>
                          <span className='truncate opacity-60'>
                            {ticket.content.slice(0, 50)}
                            {ticket.content.length > 50 ? '...' : ''}
                          </span>
                        </>
                      )}
                    </div>
                  </div>
                  <div className='flex items-center gap-2'>
                    <div className='hidden gap-2 sm:flex'>
                      <Button
                        variant='ghost'
                        size='sm'
                        onClick={() => {
                          setSelectedTicket(ticket)
                        }}
                      >
                        <Eye className='mr-1 h-4 w-4' />
                        查看
                      </Button>
                      <Button
                        variant='ghost'
                        size='sm'
                        onClick={() => handleCopy(ticket)}
                      >
                        <Copy className='mr-1 h-4 w-4' />
                        复制
                      </Button>
                    </div>
                    <div className='sm:hidden'>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant='ghost' size='icon'>
                            <MoreVertical className='h-4 w-4' />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align='end'>
                          <DropdownMenuItem
                            onClick={() => {
                              setSelectedTicket(ticket)
                            }}
                          >
                            查看详情
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleCopy(ticket)}>
                            复制内容
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>

        <div className='flex items-center justify-between px-2 py-4'>
          <p className='text-sm text-muted-foreground'>
            共 {total} 个工单，每页 {pageSize} 个
          </p>
          <div className='flex items-center gap-2'>
            <Button
              variant='outline'
              size='icon'
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              <ChevronLeft className='h-4 w-4' />
            </Button>
            <div className='flex items-center gap-1'>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                let pageNum = page
                if (totalPages <= 5) {
                  pageNum = i + 1
                } else if (page <= 3) {
                  pageNum = i + 1
                } else if (page >= totalPages - 2) {
                  pageNum = totalPages - 4 + i
                } else {
                  pageNum = page - 2 + i
                }
                return (
                  <Button
                    key={pageNum}
                    variant={page === pageNum ? 'default' : 'outline'}
                    size='sm'
                    className='h-8 w-8'
                    onClick={() => setPage(pageNum)}
                  >
                    {pageNum}
                  </Button>
                )
              })}
            </div>
            <Button
              variant='outline'
              size='icon'
              disabled={page >= totalPages}
              onClick={() => setPage(page + 1)}
            >
              <ChevronRight className='h-4 w-4' />
            </Button>
          </div>
        </div>
      </ContentLayout>

      {selectedTicket && (
        <TicketDetailDialog
          open={!!selectedTicket}
          onOpenChange={(open) => !open && setSelectedTicket(null)}
          ticket={selectedTicket}
          onSuccess={fetchTickets}
        />
      )}
    </>
  )
}

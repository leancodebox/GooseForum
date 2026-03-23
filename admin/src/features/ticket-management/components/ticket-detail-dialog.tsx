import { useEffect, useState } from 'react'
import axios from 'axios'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Ticket, TICKET_TYPES, TICKET_STATUS } from '../data/schema'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Copy, Save, Loader2 } from 'lucide-react'
import { toast } from 'sonner'

interface TicketDetailDialogProps {
  ticket: Ticket | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

export function TicketDetailDialog({
  ticket,
  open,
  onOpenChange,
  onSuccess,
}: TicketDetailDialogProps) {
  const [status, setStatus] = useState<string>('')
  const [reply, setReply] = useState<string>('')
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    if (ticket) {
      setStatus(ticket.status.toString())
      setReply(ticket.reply || '')
    }
  }, [ticket])

  if (!ticket) return null

  const handleCopy = () => {
    const content = `工单 #${ticket.id}: ${ticket.title}\n\n用户信息: ${ticket.applyUserInfo}\n类型: ${TICKET_TYPES[ticket.type as keyof typeof TICKET_TYPES]}\n状态: ${TICKET_STATUS[ticket.status as keyof typeof TICKET_STATUS]}\n\n内容:\n${ticket.content}${ticket.reply ? `\n\n回复:\n${ticket.reply}` : ''}`
    navigator.clipboard.writeText(content)
    toast.success('内容已复制到剪贴板')
  }

  const handleSave = async () => {
    setIsSaving(true)
    try {
      const response = await axios.post('/api/admin/apply-sheet-update', {
        id: ticket.id,
        status: parseInt(status),
        reply: reply,
      })
      if (response.data.code === 0) {
        toast.success('更新成功')
        onSuccess?.()
        onOpenChange(false)
      } else {
        toast.error(response.data.msg || '更新失败')
      }
    } catch (error) {
      toast.error('更新工单失败')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-[600px]'>
        <DialogHeader>
          <DialogTitle>工单详情</DialogTitle>
        </DialogHeader>
        <div className='grid gap-4 py-4 overflow-hidden'>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
            <div className='space-y-2 min-w-0'>
              <Label>工单ID</Label>
              <Input value={`#${ticket.id}`} className='w-full' readOnly />
            </div>
            <div className='space-y-2 min-w-0'>
              <Label>用户ID</Label>
              <Input value={ticket.userId} className='w-full' readOnly />
            </div>
          </div>
          <div className='space-y-2 min-w-0'>
            <Label>工单标题</Label>
            <Input value={ticket.title} className='w-full' readOnly />
          </div>
          <div className='space-y-2 min-w-0'>
            <Label>申请用户信息</Label>
            <Input value={ticket.applyUserInfo} className='w-full' readOnly />
          </div>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
            <div className='space-y-2 min-w-0'>
              <Label>工单类型</Label>
              <Input
                value={TICKET_TYPES[ticket.type as keyof typeof TICKET_TYPES]}
                className='w-full'
                readOnly
              />
            </div>
            <div className='space-y-2 min-w-0'>
              <Label>工单状态</Label>
              <Select value={status} onValueChange={setStatus}>
                <SelectTrigger className='w-full'>
                  <SelectValue placeholder='选择状态' />
                </SelectTrigger>
                <SelectContent>
                  {Object.entries(TICKET_STATUS).map(([id, label]) => (
                    <SelectItem key={id} value={id}>
                      {label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className='space-y-2 min-w-0'>
            <Label>工单内容</Label>
            <Textarea
              value={ticket.content}
              className='min-h-[100px] w-full resize-none break-all'
              readOnly
            />
          </div>
          <div className='space-y-2 min-w-0'>
            <Label>回复内容</Label>
            <Textarea
              placeholder='请输入回复内容...'
              value={reply}
              onChange={(e) => setReply(e.target.value)}
              className='min-h-[100px] w-full resize-none break-all'
            />
          </div>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
            <div className='space-y-2 min-w-0'>
              <Label>创建时间</Label>
              <Input
                value={ticket.createTime && !isNaN(Date.parse(ticket.createTime)) ? new Date(ticket.createTime).toLocaleString() : '无'}
                className='w-full'
                readOnly
              />
            </div>
            <div className='space-y-2 min-w-0'>
              <Label>更新时间</Label>
              <Input
                value={ticket.updateTime && !isNaN(Date.parse(ticket.updateTime)) ? new Date(ticket.updateTime).toLocaleString() : '无'}
                className='w-full'
                readOnly
              />
            </div>
          </div>
        </div>
        <div className='flex justify-end gap-2'>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            取消
          </Button>
          <Button variant='outline' onClick={handleCopy}>
            <Copy className='mr-2 h-4 w-4' />
            复制
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? (
              <Loader2 className='mr-2 h-4 w-4 animate-spin' />
            ) : (
              <Save className='mr-2 h-4 w-4' />
            )}
            保存
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

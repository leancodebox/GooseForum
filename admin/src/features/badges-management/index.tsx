import { useEffect, useMemo, useState } from 'react'
import { Edit3, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { ContentLayout } from '@/components/layout/content-layout'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { deleteBadge, getBadges, saveBadge } from '@/api'
import type { BadgeItem } from '@/api/types'

const emptyBadge: BadgeItem = {
  code: '',
  type: 'custom',
  grantMode: 'manual',
  name: '',
  description: '',
  iconType: 'asset',
  iconKey: '',
  iconUrl: '/static/badges/contributor.svg',
  color: 'blue',
  level: 'bronze',
  isEnabled: true,
  sortOrder: 1000,
}

const colorOptions = [
  { label: '蓝色', value: 'blue', className: 'bg-blue-100 text-blue-700 ring-blue-200' },
  { label: '翠绿', value: 'emerald', className: 'bg-emerald-100 text-emerald-700 ring-emerald-200' },
  { label: '青绿', value: 'teal', className: 'bg-teal-100 text-teal-700 ring-teal-200' },
  { label: '天蓝', value: 'sky', className: 'bg-sky-100 text-sky-700 ring-sky-200' },
  { label: '青色', value: 'cyan', className: 'bg-cyan-100 text-cyan-700 ring-cyan-200' },
  { label: '玫红', value: 'rose', className: 'bg-rose-100 text-rose-700 ring-rose-200' },
  { label: '紫罗兰', value: 'violet', className: 'bg-violet-100 text-violet-700 ring-violet-200' },
  { label: '紫色', value: 'purple', className: 'bg-purple-100 text-purple-700 ring-purple-200' },
  { label: '品红', value: 'fuchsia', className: 'bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200' },
  { label: '靛蓝', value: 'indigo', className: 'bg-indigo-100 text-indigo-700 ring-indigo-200' },
  { label: '琥珀', value: 'amber', className: 'bg-amber-100 text-amber-700 ring-amber-200' },
  { label: '橙色', value: 'orange', className: 'bg-orange-100 text-orange-700 ring-orange-200' },
  { label: '黄色', value: 'yellow', className: 'bg-yellow-100 text-yellow-700 ring-yellow-200' },
  { label: '石板灰', value: 'slate', className: 'bg-slate-100 text-slate-700 ring-slate-200' },
]

function colorLabel(value: string) {
  return colorOptions.find((item) => item.value === value)?.label || value || '蓝色'
}

function badgeIconURL(badge: BadgeItem) {
  return badge.iconUrl || '/static/badges/contributor.svg'
}

function toneClass(badge: BadgeItem) {
  if (badge.color === 'blue') return 'bg-blue-100 text-blue-700 ring-blue-200'
  if (badge.color === 'emerald') return 'bg-emerald-100 text-emerald-700 ring-emerald-200'
  if (badge.color === 'teal') return 'bg-teal-100 text-teal-700 ring-teal-200'
  if (badge.color === 'sky') return 'bg-sky-100 text-sky-700 ring-sky-200'
  if (badge.color === 'cyan') return 'bg-cyan-100 text-cyan-700 ring-cyan-200'
  if (badge.color === 'rose') return 'bg-rose-100 text-rose-700 ring-rose-200'
  if (badge.color === 'violet') return 'bg-violet-100 text-violet-700 ring-violet-200'
  if (badge.color === 'purple') return 'bg-purple-100 text-purple-700 ring-purple-200'
  if (badge.color === 'fuchsia') return 'bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-200'
  if (badge.color === 'indigo') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
  if (badge.color === 'amber') return 'bg-amber-100 text-amber-700 ring-amber-200'
  if (badge.color === 'orange') return 'bg-orange-100 text-orange-700 ring-orange-200'
  if (badge.color === 'yellow') return 'bg-yellow-100 text-yellow-700 ring-yellow-200'
  if (badge.color === 'slate') return 'bg-slate-100 text-slate-700 ring-slate-200'
  if (badge.level === 'gold') return 'bg-amber-100 text-amber-700 ring-amber-200'
  if (badge.level === 'special') return 'bg-indigo-100 text-indigo-700 ring-indigo-200'
  return 'bg-blue-100 text-blue-700 ring-blue-200'
}

export default function BadgesManagement() {
  const [badges, setBadges] = useState<BadgeItem[]>([])
  const [editing, setEditing] = useState<BadgeItem | null>(null)
  const [saving, setSaving] = useState(false)
  const [loading, setLoading] = useState(false)

  const stats = useMemo(() => ({
    system: badges.filter((item) => item.type === 'system').length,
    custom: badges.filter((item) => item.type === 'custom').length,
  }), [badges])

  async function loadBadges() {
    setLoading(true)
    try {
      const response = await getBadges()
      if (response.code === 0) setBadges(response.result || [])
      else toast.error(response.msg || '加载徽章失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadBadges()
  }, [])

  async function submitBadge() {
    if (!editing) return
    if (!editing.name.trim()) {
      toast.error('徽章名称不能为空')
      return
    }
    setSaving(true)
    try {
      const payload = {
        ...editing,
        code: editing.code.trim(),
        name: editing.name.trim(),
        type: editing.type || 'custom',
        grantMode: editing.grantMode || 'manual',
        iconType: editing.iconType || 'asset',
        color: colorOptions.some((item) => item.value === editing.color) ? editing.color : 'blue',
      }
      const response = await saveBadge(payload)
      if (response.code === 0) {
        toast.success('保存成功')
        setEditing(null)
        await loadBadges()
      } else {
        toast.error(response.msg || '保存失败')
      }
    } finally {
      setSaving(false)
    }
  }

  async function removeBadge(badge: BadgeItem) {
    if (badge.type === 'system') return
    if (!window.confirm(`确认删除徽章「${badge.name}」吗？`)) return
    const response = await deleteBadge(badge.code)
    if (response.code === 0) {
      toast.success('已删除')
      await loadBadges()
    } else {
      toast.error(response.msg || '删除失败')
    }
  }

  return (
    <>
      <ContentLayout
        title='徽章管理'
        description='管理系统默认徽章和自定义徽章。系统默认徽章只能编辑，不能删除。'
        headerActions={
          <Button onClick={() => setEditing({ ...emptyBadge })}>
            <Plus className='mr-2 h-4 w-4' />
            新增徽章
          </Button>
        }
      >
        <div className='mb-3 flex flex-wrap gap-2 text-sm text-muted-foreground'>
          <Badge variant='secondary'>系统默认 {stats.system}</Badge>
          <Badge variant='outline'>自定义 {stats.custom}</Badge>
          {loading && <span>加载中...</span>}
        </div>

        <div className='grid grid-cols-[repeat(auto-fill,minmax(5.5rem,1fr))] gap-3'>
          {badges.map((badge) => (
            <div key={badge.code} className='group relative flex min-w-0 flex-col items-center rounded-md px-1 py-1.5 transition-colors hover:bg-muted/60'>
              <button
                type='button'
                className='flex min-w-0 flex-col items-center'
                title={`${badge.name} · ${badge.code}`}
                onClick={() => setEditing({ ...badge })}
              >
                  <div
                    className={`flex h-12 w-12 shrink-0 items-center justify-center ring-1 ring-inset transition-transform group-hover:scale-105 ${toneClass(badge)}`}
                    style={{ clipPath: 'polygon(25% 5%, 75% 5%, 100% 50%, 75% 95%, 25% 95%, 0 50%)' }}
                  >
                    <img src={badgeIconURL(badge)} alt={badge.name} className='h-6 w-6 object-contain' />
                  </div>
                  <div className='mt-1 flex max-w-full items-center gap-1'>
                    <span className={`truncate text-xs font-semibold leading-5 ${badge.type === 'custom' ? 'text-blue-600' : 'text-foreground'}`}>
                      {badge.name}
                    </span>
                  </div>
                  <div className='max-w-full truncate text-[10px] leading-4 text-muted-foreground'>{badge.grantMode === 'auto' ? '自动' : '手动'} · {badge.isEnabled ? badge.level || 'bronze' : '停用'}</div>
              </button>
              <div className='absolute right-0.5 top-0.5 flex gap-0.5 rounded-md bg-background/90 p-0.5 opacity-0 shadow-sm ring-1 ring-border transition-opacity group-hover:opacity-100'>
                <Button variant='ghost' size='icon' className='h-6 w-6' title='编辑' onClick={() => setEditing({ ...badge })}>
                  <Edit3 className='h-3.5 w-3.5' />
                </Button>
                {badge.type !== 'system' && (
                  <Button variant='ghost' size='icon' className='h-6 w-6 text-destructive hover:text-destructive' title='删除' onClick={() => void removeBadge(badge)}>
                    <Trash2 className='h-3.5 w-3.5' />
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      </ContentLayout>

      <Dialog open={Boolean(editing)} onOpenChange={(open) => !open && setEditing(null)}>
        <DialogContent className='sm:max-w-2xl'>
          <DialogHeader>
            <DialogTitle>{editing?.code ? '编辑徽章' : '新增徽章'}</DialogTitle>
            <DialogDescription>系统默认徽章会保存为覆盖配置，自定义徽章会保存为独立徽章。</DialogDescription>
          </DialogHeader>
          {editing && (
            <div className='grid max-h-[68vh] gap-4 overflow-y-auto pr-2 sm:grid-cols-2'>
              <div className='space-y-2'>
                <Label>编码</Label>
                <Input value={editing.code || '保存后自动生成'} disabled />
              </div>
              <div className='space-y-2'>
                <Label>名称</Label>
                <Input value={editing.name} onChange={(event) => setEditing({ ...editing, name: event.target.value })} />
              </div>
              <div className='space-y-2 sm:col-span-2'>
                <Label>描述</Label>
                <Textarea value={editing.description} onChange={(event) => setEditing({ ...editing, description: event.target.value })} />
              </div>
              <div className='space-y-2'>
                <Label>图标 URL</Label>
                <Input value={editing.iconUrl} onChange={(event) => setEditing({ ...editing, iconUrl: event.target.value })} />
              </div>
              <div className='space-y-2'>
                <Label>颜色</Label>
                <div className='flex gap-2'>
                  <select
                    className='h-9 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm'
                    value={colorOptions.some((item) => item.value === editing.color) ? editing.color : 'blue'}
                    onChange={(event) => setEditing({ ...editing, color: event.target.value })}
                  >
                    {colorOptions.map((color) => (
                      <option key={color.value} value={color.value}>{color.label}</option>
                    ))}
                  </select>
                  <div className={`flex h-9 min-w-20 items-center justify-center rounded-md px-2 text-xs font-semibold ring-1 ring-inset ${toneClass(editing)}`}>
                    {colorLabel(editing.color)}
                  </div>
                </div>
              </div>
              <div className='space-y-2'>
                <Label>等级</Label>
                <Input value={editing.level} onChange={(event) => setEditing({ ...editing, level: event.target.value })} />
              </div>
              <div className='space-y-2'>
                <Label>排序</Label>
                <Input type='number' value={editing.sortOrder} onChange={(event) => setEditing({ ...editing, sortOrder: Number(event.target.value) })} />
              </div>
              <div className='space-y-2'>
                <Label>授予方式</Label>
                <select
                  className='h-9 w-full rounded-md border bg-background px-3 text-sm disabled:cursor-not-allowed disabled:opacity-60'
                  value={editing.grantMode}
                  disabled={editing.type === 'system'}
                  onChange={(event) => setEditing({ ...editing, grantMode: event.target.value as BadgeItem['grantMode'] })}
                >
                  <option value='auto'>自动</option>
                  <option value='manual'>手动</option>
                </select>
                {editing.type === 'system' && (
                  <div className='text-xs text-muted-foreground'>系统默认徽章的授予方式由代码逻辑决定。</div>
                )}
              </div>
              <div className='space-y-2'>
                <Label>启用徽章</Label>
                <div className='flex h-9 items-center justify-between rounded-md border bg-background px-3'>
                  <span className='text-sm text-muted-foreground'>停用后不展示也不可授予</span>
                  <Switch checked={editing.isEnabled} onCheckedChange={(checked) => setEditing({ ...editing, isEnabled: checked })} />
                </div>
              </div>
            </div>
          )}
          <DialogFooter>
            <Button variant='outline' onClick={() => setEditing(null)}>取消</Button>
            <Button disabled={saving} onClick={() => void submitBadge()}>{saving ? '保存中...' : '保存'}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}

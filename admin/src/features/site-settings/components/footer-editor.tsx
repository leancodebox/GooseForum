import { useState } from 'react'
import { useFieldArray, type Control } from 'react-hook-form'
import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd'
import { Plus, GripVertical, Trash2, Edit } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import { type SiteInfo } from '../data/schema'

interface FooterEditorProps {
  control: Control<SiteInfo>
}

export function FooterEditor({ control }: FooterEditorProps) {
  const {
    fields: primaryFields,
    append: appendPrimary,
    remove: removePrimary,
    update: updatePrimary,
    move: movePrimary,
  } = useFieldArray({
    control,
    name: 'footerInfo.primary',
  })

  const {
    fields: listFields,
    append: appendList,
    remove: removeList,
    update: updateList,
    move: moveList,
  } = useFieldArray({
    control,
    name: 'footerInfo.list',
  })

  const [editIndex, setEditIndex] = useState<number | null>(null)
  const [editType, setEditType] = useState<'primary' | 'list' | null>(null)
  const [editData, setEditData] = useState<any>(null)

  const handleDragEnd = (result: DropResult) => {
    if (!result.destination) return

    const { source, destination, type } = result
    if (source.index === destination.index) return

    if (type === 'primary') {
      movePrimary(source.index, destination.index)
    } else if (type === 'list') {
      moveList(source.index, destination.index)
    }
  }

  const openEdit = (type: 'primary' | 'list', index: number) => {
    setEditType(type)
    setEditIndex(index)
    if (type === 'primary') {
      setEditData({ ...primaryFields[index] })
    } else {
      setEditData({ ...listFields[index] })
    }
  }

  const saveEdit = () => {
    if (editType === 'primary' && editIndex !== null) {
      updatePrimary(editIndex, { content: editData.content })
    } else if (editType === 'list' && editIndex !== null) {
      updateList(editIndex, { name: editData.name, url: editData.url })
    }
    setEditIndex(null)
    setEditType(null)
  }

  return (
    <div>
      <DragDropContext onDragEnd={handleDragEnd}>
        <div>
          <div className='space-y-2 border-b border-border/70 pb-3'>
            <div className='flex items-center justify-between'>
              <div className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'>第一行：链接</div>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='h-8 px-2'
                onClick={() => appendList({ name: '新链接', url: '#' })}
              >
                <Plus className='mr-1 h-4 w-4' /> 添加链接
              </Button>
            </div>
            <Droppable droppableId='footer-list' type='list' direction='horizontal'>
              {(provided) => (
                <div
                  {...provided.droppableProps}
                  ref={provided.innerRef}
                  className='flex min-h-9 flex-wrap items-center gap-x-3 gap-y-1.5'
                >
                  {listFields.map((field, index) => (
                    <Draggable key={field.id} draggableId={field.id} index={index}>
                      {(provided) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          className='group -ml-1 inline-flex items-center rounded-md px-1 py-1 text-sm text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground focus-within:bg-muted/60'
                        >
                          <div
                            {...provided.dragHandleProps}
                            className='mr-0.5 cursor-grab text-muted-foreground/50 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100 active:cursor-grabbing'
                          >
                            <GripVertical className='h-3.5 w-3.5' />
                          </div>
                          <button
                            type='button'
                            className='min-h-6 max-w-[11rem] truncate rounded-sm px-1 text-left font-medium outline-none hover:text-primary focus-visible:ring-1 focus-visible:ring-ring'
                            onClick={() => openEdit('list', index)}
                          >
                            {field.name || '未命名'}
                          </button>
                          <div className='ml-0.5 flex items-center opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100'>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-6 w-6 rounded-sm hover:bg-background'
                              onClick={() => openEdit('list', index)}
                            >
                              <Edit className='h-3.5 w-3.5' />
                            </Button>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-6 w-6 rounded-sm hover:bg-destructive/10 hover:text-destructive'
                              onClick={() => removeList(index)}
                            >
                              <Trash2 className='h-3.5 w-3.5' />
                            </Button>
                          </div>
                        </div>
                      )}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                  {listFields.length === 0 && (
                    <div className='py-1 text-sm text-muted-foreground'>
                      暂无链接
                    </div>
                  )}
                </div>
              )}
            </Droppable>
          </div>

          {/* Primary 区域 */}
          <div className='space-y-2 pt-3'>
            <div className='flex items-center justify-between'>
              <div className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'>第二行：文字内容</div>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                className='h-8 px-2'
                onClick={() => appendPrimary({ content: '新内容' })}
              >
                <Plus className='mr-1 h-4 w-4' /> 添加文字
              </Button>
            </div>
            <Droppable droppableId='footer-primary' type='primary'>
              {(provided) => (
                <div
                  {...provided.droppableProps}
                  ref={provided.innerRef}
                  className='flex min-h-9 flex-wrap items-center gap-x-3 gap-y-1.5'
                >
                  {primaryFields.map((field, index) => (
                    <Draggable key={field.id} draggableId={field.id} index={index}>
                      {(provided) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          className='group -ml-1 inline-flex max-w-full items-center rounded-md px-1 py-1 text-sm text-muted-foreground transition-colors hover:bg-muted/60 hover:text-foreground focus-within:bg-muted/60'
                        >
                          <div
                            {...provided.dragHandleProps}
                            className='mr-0.5 cursor-grab text-muted-foreground/50 opacity-0 transition-opacity hover:text-foreground group-hover:opacity-100 active:cursor-grabbing'
                          >
                            <GripVertical className='h-3.5 w-3.5' />
                          </div>
                          <button
                            type='button'
                            className='min-h-6 max-w-[20rem] truncate rounded-sm px-1 text-left font-medium outline-none hover:text-primary focus-visible:ring-1 focus-visible:ring-ring'
                            onClick={() => openEdit('primary', index)}
                          >
                            {field.content || '空内容'}
                          </button>
                          <div className='ml-0.5 flex items-center opacity-0 transition-opacity group-hover:opacity-100 group-focus-within:opacity-100'>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-6 w-6 rounded-sm hover:bg-background'
                              onClick={() => openEdit('primary', index)}
                            >
                              <Edit className='h-3.5 w-3.5' />
                            </Button>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-6 w-6 rounded-sm hover:bg-destructive/10 hover:text-destructive'
                              onClick={() => removePrimary(index)}
                            >
                              <Trash2 className='h-3.5 w-3.5' />
                            </Button>
                          </div>
                        </div>
                      )}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                  {primaryFields.length === 0 && (
                    <div className='py-1 text-sm text-muted-foreground'>
                      暂无内容
                    </div>
                  )}
                </div>
              )}
            </Droppable>
          </div>
        </div>
      </DragDropContext>

      {/* 编辑弹窗 */}
      <Dialog open={editIndex !== null} onOpenChange={(open) => !open && setEditIndex(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              编辑 {editType === 'primary' ? '文字内容' : '链接项'}
            </DialogTitle>
          </DialogHeader>
          <div className='space-y-4 py-4'>
            {editType === 'list' && (
              <>
                <div className='space-y-2'>
                  <Label>显示名称</Label>
                  <Input
                    value={editData?.name || ''}
                    onChange={(e) => setEditData({ ...editData, name: e.target.value })}
                    placeholder='例如: GitHub'
                  />
                </div>
                <div className='space-y-2'>
                  <Label>链接地址 URL</Label>
                  <Input
                    value={editData?.url || ''}
                    onChange={(e) => setEditData({ ...editData, url: e.target.value })}
                    placeholder='例如: https://github.com'
                  />
                </div>
              </>
            )}
            {editType === 'primary' && (
              <div className='space-y-2'>
                <Label>文字内容</Label>
                <Input
                  value={editData?.content || ''}
                  onChange={(e) => setEditData({ ...editData, content: e.target.value })}
                  placeholder='例如: GooseForum © 2024'
                />
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant='outline' onClick={() => setEditIndex(null)}>
              取消
            </Button>
            <Button onClick={saveEdit}>保存</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

import { useState } from 'react'
import { useFieldArray, type Control } from 'react-hook-form'
import { DragDropContext, Droppable, Draggable, type DropResult } from '@hello-pangea/dnd'
import { Plus, GripVertical, Trash2, Edit, Type } from 'lucide-react'
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
    <div className='space-y-8'>
      <DragDropContext onDragEnd={handleDragEnd}>
        <div className='grid gap-8 md:grid-cols-2'>
          <div className='space-y-4'>
            <div className='flex items-center justify-between'>
              <div className='text-sm font-medium text-muted-foreground'>上方链接列表 (List)</div>
              <Button
                type='button'
                variant='outline'
                size='sm'
                onClick={() => appendList({ name: '新链接', url: '#' })}
              >
                <Plus className='h-4 w-4 mr-1' /> 添加链接
              </Button>
            </div>
            <Droppable droppableId='footer-list' type='list' direction='horizontal'>
              {(provided) => (
                <div
                  {...provided.droppableProps}
                  ref={provided.innerRef}
                  className='flex flex-wrap gap-2 min-h-[48px] p-2 bg-muted/30 rounded-lg border border-dashed'
                >
                  {listFields.map((field, index) => (
                    <Draggable key={field.id} draggableId={field.id} index={index}>
                      {(provided) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          className='flex items-center gap-1.5 pl-1.5 pr-1 py-1 bg-background rounded-md border shadow-sm group hover:border-primary/50 transition-colors'
                        >
                          <div
                            {...provided.dragHandleProps}
                            className='cursor-grab text-muted-foreground hover:text-foreground active:cursor-grabbing'
                          >
                            <GripVertical className='h-3.5 w-3.5' />
                          </div>
                          <span
                            className='text-xs font-medium cursor-pointer hover:underline'
                            onClick={() => openEdit('list', index)}
                          >
                            {field.name || '未命名'}
                          </span>
                          <div className='flex items-center ml-1'>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-5 w-5 rounded-sm hover:bg-destructive/10 hover:text-destructive'
                              onClick={() => removeList(index)}
                            >
                              <Trash2 className='h-3 w-3' />
                            </Button>
                          </div>
                        </div>
                      )}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                  {listFields.length === 0 && (
                    <div className='w-full text-center text-sm text-muted-foreground py-2'>
                      暂无链接
                    </div>
                  )}
                </div>
              )}
            </Droppable>
          </div>

          {/* Primary 区域 */}
          <div className='space-y-4'>
            <div className='flex items-center justify-between'>
              <div className='text-sm font-medium text-muted-foreground'>下方文字内容 (Primary)</div>
              <Button
                type='button'
                variant='outline'
                size='sm'
                onClick={() => appendPrimary({ content: '新内容' })}
              >
                <Plus className='h-4 w-4 mr-1' /> 添加文字
              </Button>
            </div>
            <Droppable droppableId='footer-primary' type='primary'>
              {(provided) => (
                <div
                  {...provided.droppableProps}
                  ref={provided.innerRef}
                  className='space-y-2'
                >
                  {primaryFields.map((field, index) => (
                    <Draggable key={field.id} draggableId={field.id} index={index}>
                      {(provided) => (
                        <div
                          ref={provided.innerRef}
                          {...provided.draggableProps}
                          className='flex items-center justify-between p-2 md:p-3 bg-muted/50 rounded-lg border group'
                        >
                          <div className='flex items-center gap-2 overflow-hidden'>
                            <div
                              {...provided.dragHandleProps}
                              className='cursor-grab text-muted-foreground hover:text-foreground'
                            >
                              <GripVertical className='h-4 w-4' />
                            </div>
                            <Type className='h-4 w-4 text-muted-foreground shrink-0' />
                            <span className='text-sm truncate'>
                              {field.content || '空内容'}
                            </span>
                          </div>
                          <div className='flex items-center gap-1 opacity-100 md:opacity-0 md:group-hover:opacity-100 transition-opacity'>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-8 w-8'
                              onClick={() => openEdit('primary', index)}
                            >
                              <Edit className='h-4 w-4' />
                            </Button>
                            <Button
                              type='button'
                              variant='ghost'
                              size='icon'
                              className='h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10'
                              onClick={() => removePrimary(index)}
                            >
                              <Trash2 className='h-4 w-4' />
                            </Button>
                          </div>
                        </div>
                      )}
                    </Draggable>
                  ))}
                  {provided.placeholder}
                  {primaryFields.length === 0 && (
                    <div className='text-center p-4 border border-dashed rounded-lg text-sm text-muted-foreground'>
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
                <Label>文字内容 / HTML</Label>
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

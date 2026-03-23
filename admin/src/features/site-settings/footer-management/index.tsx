import { useEffect, useState } from 'react'
import {
  DragDropContext,
  Droppable,
  Draggable,
  type DropResult,
} from '@hello-pangea/dnd'
import { Plus, GripVertical, Trash2, Save, Type, Link as LinkIcon, Code } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ContentLayout } from '@/components/layout/content-layout'
import { Input } from '@/components/ui/input'
import axios from 'axios'
import { type FooterConfig, type FooterItem } from './data/schema'

export default function FooterManagement() {
  const [config, setConfig] = useState<FooterConfig>({
    primary: [],
    list: []
  })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const fetchConfig = async () => {
    try {
      const res = await axios.get('/api/admin/footer-links')
      if (res.data.code === 0) {
        setConfig({
          primary: res.data.result?.primary || [],
          list: res.data.result?.list || []
        })
      } else {
        toast.error(res.data.msg || '获取页脚配置失败')
      }
    } catch (error) {
      toast.error('获取页脚配置失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchConfig()
  }, [])

  const saveConfig = async (newConfig: FooterConfig) => {
    setSaving(true)
    try {
      const res = await axios.post('/api/admin/save-footer-links', {
        footerConfig: newConfig,
      })
      if (res.data.code === 0) {
        toast.success('保存成功')
        setConfig(newConfig)
      } else {
        toast.error(res.data.msg || '保存失败')
      }
    } catch (error) {
      toast.error('保存失败')
    } finally {
      setSaving(false)
    }
  }

  const handleSave = () => {
    saveConfig(config)
  }

  const onDragEnd = (result: DropResult) => {
    const { source, destination, type } = result

    if (!destination) return

    if (
      source.droppableId === destination.droppableId &&
      source.index === destination.index
    ) {
      return
    }

    const newConfig = { ...config }

    if (type === 'primary') {
      const [removed] = newConfig.primary.splice(source.index, 1)
      newConfig.primary.splice(destination.index, 0, removed)
    } else if (type === 'group') {
      const [removed] = newConfig.list.splice(source.index, 1)
      newConfig.list.splice(destination.index, 0, removed)
    } else if (type.startsWith('child-')) {
      const groupIdx = parseInt(type.split('-')[1])
      const [removed] = newConfig.list[groupIdx].children.splice(source.index, 1)
      newConfig.list[groupIdx].children.splice(destination.index, 0, removed)
    }

    setConfig(newConfig)
  }

  const addPItem = () => {
    const newConfig = { ...config }
    newConfig.primary.push({ content: '' })
    setConfig(newConfig)
  }

  const removePItem = (index: number) => {
    const newConfig = { ...config }
    newConfig.primary.splice(index, 1)
    setConfig(newConfig)
  }

  const updatePItem = (index: number, content: string) => {
    const newConfig = { ...config }
    newConfig.primary[index].content = content
    setConfig(newConfig)
  }

  const addGroup = () => {
    const newConfig = { ...config }
    newConfig.list.push({ name: '', children: [] })
    setConfig(newConfig)
  }

  const removeGroup = (index: number) => {
    const newConfig = { ...config }
    newConfig.list.splice(index, 1)
    setConfig(newConfig)
  }

  const updateGroupName = (index: number, name: string) => {
    const newConfig = { ...config }
    newConfig.list[index].name = name
    setConfig(newConfig)
  }

  const addChild = (groupIdx: number) => {
    const newConfig = { ...config }
    newConfig.list[groupIdx].children.push({ name: '', url: '' })
    setConfig(newConfig)
  }

  const removeChild = (groupIdx: number, childIdx: number) => {
    const newConfig = { ...config }
    newConfig.list[groupIdx].children.splice(childIdx, 1)
    setConfig(newConfig)
  }

  const updateChild = (groupIdx: number, childIdx: number, field: keyof FooterItem, value: string) => {
    const newConfig = { ...config }
    newConfig.list[groupIdx].children[childIdx][field] = value
    setConfig(newConfig)
  }

  return (
    <>
      <ContentLayout
        title='页脚管理'
        description='管理论坛底部的文字内容和链接列表。'
        headerActions={
          <Button onClick={handleSave} disabled={saving}>
            <Save className='mr-2 h-4 w-4' /> {saving ? '保存中...' : '保存配置'}
          </Button>
        }
      >
        {loading ? (
          <div className='flex h-64 items-center justify-center'>
            <p className='text-muted-foreground'>加载中...</p>
          </div>
        ) : (
          <DragDropContext onDragEnd={onDragEnd}>
            <div className='grid grid-cols-1 lg:grid-cols-2 gap-4'>
              {/* HTML内容部分 */}
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-lg font-medium'>
                    <div className='flex items-center gap-2'>
                      <Code className='h-5 w-5' />
                      HTML内容列表
                    </div>
                  </CardTitle>
                  <Button variant='outline' size='sm' onClick={addPItem}>
                    <Plus className='h-4 w-4 mr-1' /> 添加HTML项
                  </Button>
                </CardHeader>
                <CardContent className='pt-2'>
                  <Droppable droppableId='primary' type='primary'>
                    {(provided) => (
                      <div
                        {...provided.droppableProps}
                        ref={provided.innerRef}
                        className='space-y-3'
                      >
                        {config.primary.map((item, index) => (
                          <Draggable
                            key={`primary-${index}`}
                            draggableId={`primary-${index}`}
                            index={index}
                          >
                            {(provided) => (
                              <div
                                ref={provided.innerRef}
                                {...provided.draggableProps}
                                className='flex items-center gap-2 bg-secondary/30 p-2 rounded-md group'
                              >
                                <div
                                  {...provided.dragHandleProps}
                                  className='cursor-move text-muted-foreground hover:text-foreground'
                                >
                                  <GripVertical className='h-4 w-4' />
                                </div>
                                <Input
                                  value={item.content}
                                  onChange={(e) => updatePItem(index, e.target.value)}
                                  placeholder='请输入HTML项内容'
                                  className='flex-1 h-9'
                                />
                                <Button
                                  variant='ghost'
                                  size='icon'
                                  className='h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10'
                                  onClick={() => removePItem(index)}
                                >
                                  <Trash2 className='h-4 w-4' />
                                </Button>
                              </div>
                            )}
                          </Draggable>
                        ))}
                        {provided.placeholder}
                        {config.primary.length === 0 && (
                          <div className='text-center py-8 border-2 border-dashed rounded-md text-muted-foreground'>
                            暂无HTML项，点击上方按钮添加
                          </div>
                        )}
                      </div>
                    )}
                  </Droppable>
                </CardContent>
              </Card>

              {/* 链接分组部分 */}
              <Card>
                <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
                  <CardTitle className='text-lg font-medium'>
                    <div className='flex items-center gap-2'>
                      <Type className='h-5 w-5' />
                      链接分组
                    </div>
                  </CardTitle>
                  <Button variant='outline' size='sm' onClick={addGroup}>
                    <Plus className='h-4 w-4 mr-1' /> 添加分组
                  </Button>
                </CardHeader>
                <CardContent className='pt-2'>
                  <Droppable droppableId='groups' type='group'>
                    {(provided) => (
                      <div
                        {...provided.droppableProps}
                        ref={provided.innerRef}
                        className='space-y-4'
                      >
                        {config.list.map((group, groupIdx) => (
                          <Draggable
                            key={`group-${groupIdx}`}
                            draggableId={`group-${groupIdx}`}
                            index={groupIdx}
                          >
                            {(provided) => (
                              <div
                                ref={provided.innerRef}
                                {...provided.draggableProps}
                                className='bg-secondary/30 p-3 rounded-md border border-border'
                              >
                                <div className='flex items-center justify-between mb-3 gap-2'>
                                  <div className='flex items-center gap-2 flex-1'>
                                    <div
                                      {...provided.dragHandleProps}
                                      className='cursor-move text-muted-foreground hover:text-foreground'
                                    >
                                      <GripVertical className='h-4 w-4' />
                                    </div>
                                    <Input
                                      value={group.name}
                                      onChange={(e) => updateGroupName(groupIdx, e.target.value)}
                                      placeholder='分组名称'
                                      className='h-9 font-medium'
                                    />
                                  </div>
                                  <Button
                                    variant='ghost'
                                    size='icon'
                                    className='h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10'
                                    onClick={() => removeGroup(groupIdx)}
                                  >
                                    <Trash2 className='h-4 w-4' />
                                  </Button>
                                </div>

                                <Droppable droppableId={`group-${groupIdx}-children`} type={`child-${groupIdx}`}>
                                  {(provided) => (
                                    <div
                                      {...provided.droppableProps}
                                      ref={provided.innerRef}
                                      className='space-y-2 ml-6'
                                    >
                                      {group.children.map((child, childIdx) => (
                                        <Draggable
                                          key={`child-${groupIdx}-${childIdx}`}
                                          draggableId={`child-${groupIdx}-${childIdx}`}
                                          index={childIdx}
                                        >
                                          {(provided) => (
                                            <div
                                              ref={provided.innerRef}
                                              {...provided.draggableProps}
                                              className='flex items-center gap-2 group'
                                            >
                                              <div
                                                {...provided.dragHandleProps}
                                                className='cursor-move text-muted-foreground/40 hover:text-foreground'
                                              >
                                                <GripVertical className='h-3 w-3' />
                                              </div>
                                              <div className='flex-1 grid grid-cols-2 gap-2'>
                                                <div className='relative'>
                                                  <Type className='absolute left-2 top-2.5 h-3 w-3 text-muted-foreground' />
                                                  <Input
                                                    value={child.name}
                                                    onChange={(e) => updateChild(groupIdx, childIdx, 'name', e.target.value)}
                                                    placeholder='名称'
                                                    className='h-8 pl-7 text-xs'
                                                  />
                                                </div>
                                                <div className='relative'>
                                                  <LinkIcon className='absolute left-2 top-2.5 h-3 w-3 text-muted-foreground' />
                                                  <Input
                                                    value={child.url}
                                                    onChange={(e) => updateChild(groupIdx, childIdx, 'url', e.target.value)}
                                                    placeholder='URL'
                                                    className='h-8 pl-7 text-xs'
                                                  />
                                                </div>
                                              </div>
                                              <Button
                                                variant='ghost'
                                                size='icon'
                                                className='h-7 w-7 text-destructive hover:text-destructive hover:bg-destructive/10'
                                                onClick={() => removeChild(groupIdx, childIdx)}
                                              >
                                                <Trash2 className='h-3.5 w-3.5' />
                                              </Button>
                                            </div>
                                          )}
                                        </Draggable>
                                      ))}
                                      {provided.placeholder}
                                      <Button
                                        variant='outline'
                                        size='sm'
                                        className='w-full h-8 border-dashed mt-2'
                                        onClick={() => addChild(groupIdx)}
                                      >
                                        <Plus className='h-3 w-3 mr-1' /> 添加链接
                                      </Button>
                                    </div>
                                  )}
                                </Droppable>
                              </div>
                            )}
                          </Draggable>
                        ))}
                        {provided.placeholder}
                        {config.list.length === 0 && (
                          <div className='text-center py-8 border-2 border-dashed rounded-md text-muted-foreground'>
                            暂无分组，点击上方按钮添加
                          </div>
                        )}
                      </div>
                    )}
                  </Droppable>
                </CardContent>
              </Card>
            </div>
          </DragDropContext>
        )}
      </ContentLayout>
    </>
  )
}

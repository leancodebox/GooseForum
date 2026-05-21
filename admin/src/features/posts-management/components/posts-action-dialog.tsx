import { useEffect, useState } from 'react'
import { toast } from 'sonner'
import { editArticle, editArticleCategories, getArticleSource, getCategoryList } from '@/api'
import type { ArticleSource, Category } from '@/api/types'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { usePosts } from './posts-provider'

export function PostsActionDialog() {
  const { open, setOpen, currentRow, setCurrentRow } = usePosts()
  const [categories, setCategories] = useState<Category[]>([])
  const [categoryLoading, setCategoryLoading] = useState(false)
  const [categoryLoaded, setCategoryLoaded] = useState(false)
  const [categorySaving, setCategorySaving] = useState(false)
  const [selectedCategoryIds, setSelectedCategoryIds] = useState<number[]>([])
  const [source, setSource] = useState<ArticleSource | null>(null)
  const [sourceLoading, setSourceLoading] = useState(false)
  const isCategoryDialog = open === 'categories'
  const isSourceDialog = open === 'source'

  useEffect(() => {
    if (!isCategoryDialog || !currentRow) return
    setSelectedCategoryIds(currentRow.categoryId || [])
    if (categoryLoaded || categoryLoading) return

    setCategoryLoading(true)
    getCategoryList()
      .then((res) => {
        if (res.code === 0) {
          setCategories(res.result || [])
        } else {
          toast.error(res.msg || '分类加载失败')
        }
      })
      .catch(() => {
        toast.error('分类加载失败')
      })
      .finally(() => {
        setCategoryLoaded(true)
        setCategoryLoading(false)
      })
  }, [categoryLoaded, categoryLoading, currentRow, isCategoryDialog])

  useEffect(() => {
    if (!isSourceDialog || !currentRow) return

    setSource(null)
    setSourceLoading(true)
    getArticleSource(currentRow.id)
      .then((res) => {
        if (res.code === 0) {
          setSource(res.result || null)
        } else {
          toast.error(res.msg || '原文加载失败')
        }
      })
      .catch(() => {
        toast.error('原文加载失败')
      })
      .finally(() => {
        setSourceLoading(false)
      })
  }, [currentRow, isSourceDialog])

  const closeDialog = () => {
    setOpen(null)
    setCurrentRow(null)
    setSource(null)
  }

  const copySource = async () => {
    if (!source?.content) return
    try {
      await navigator.clipboard.writeText(source.content)
      toast.success('原文已复制')
    } catch {
      toast.error('复制失败')
    }
  }

  const handleAction = async () => {
    if (!currentRow) return

    try {
      if (open === 'top' || open === 'recommend') {
        toast.info('该功能暂未实现')
        closeDialog()
        return
      }

      const response = await editArticle({
        id: currentRow.id,
        processStatus: open === 'reject' ? 1 : 0,
      })

      if (response.code === 0) {
        toast.success('操作成功')
        closeDialog()
        window.location.reload()
      } else {
        toast.error(response.msg || '操作失败')
      }
    } catch {
      toast.error('请求失败')
    }
  }

  const toggleCategory = (categoryId: number) => {
    setSelectedCategoryIds((prev) => {
      if (prev.includes(categoryId)) {
        return prev.filter((id) => id !== categoryId)
      }
      if (prev.length >= 3) {
        toast.warning('最多选择 3 个分类')
        return prev
      }
      return [...prev, categoryId]
    })
  }

  const handleSaveCategories = async () => {
    if (!currentRow || categorySaving) return
    if (selectedCategoryIds.length === 0) {
      toast.warning('至少选择一个分类')
      return
    }

    setCategorySaving(true)
    try {
      const response = await editArticleCategories({
        id: currentRow.id,
        categoryId: selectedCategoryIds,
      })
      if (response.code === 0) {
        toast.success('分类已更新')
        closeDialog()
        window.location.reload()
      } else {
        toast.error(response.msg || '分类更新失败')
      }
    } catch {
      toast.error('分类更新失败')
    } finally {
      setCategorySaving(false)
    }
  }

  const getTitle = () => {
    switch (open) {
      case 'approve': return '确认恢复文章？'
      case 'reject': return '确认封禁文章？'
      case 'top': return '确认置顶文章？'
      case 'recommend': return '确认推荐文章？'
      default: return ''
    }
  }

  const getDescription = () => {
    switch (open) {
      case 'approve': return `确认要恢复文章 "${currentRow?.title}" 吗？`
      case 'reject': return `确认要封禁文章 "${currentRow?.title}" 吗？封禁后用户将无法查看。`
      case 'top': return `确认要置顶文章 "${currentRow?.title}" 吗？`
      case 'recommend': return `确认要推荐文章 "${currentRow?.title}" 吗？`
      default: return ''
    }
  }

  return (
    <>
      <Dialog
        open={isCategoryDialog}
        onOpenChange={(val) => !val && closeDialog()}
      >
        <DialogContent className='sm:max-w-[560px]'>
          <DialogHeader>
            <DialogTitle>修改文章分类</DialogTitle>
            <DialogDescription className='line-clamp-2'>
              为「{currentRow?.title}」选择 1 到 3 个分类。
            </DialogDescription>
          </DialogHeader>

          <div className='rounded-lg border bg-muted/20 p-3'>
            {categoryLoading ? (
              <div className='py-8 text-center text-sm text-muted-foreground'>
                正在加载分类...
              </div>
            ) : categories.length === 0 ? (
              <div className='py-8 text-center text-sm text-muted-foreground'>
                暂无可用分类
              </div>
            ) : (
              <div className='grid gap-2 sm:grid-cols-2'>
                {categories.map((category) => {
                  const checked = selectedCategoryIds.includes(category.id)
                  return (
                    <button
                      key={category.id}
                      type='button'
                      onClick={() => toggleCategory(category.id)}
                      className={cn(
                        'flex min-h-12 items-center gap-3 rounded-md border bg-background px-3 py-2 text-left text-sm transition-colors hover:border-primary/50 hover:bg-primary/5',
                        checked && 'border-primary bg-primary/10 text-primary'
                      )}
                    >
                      <span
                        className='size-2.5 shrink-0 rounded-full'
                        style={{ backgroundColor: category.color || '#64748b' }}
                      />
                      <span className='min-w-0 flex-1 truncate font-medium'>
                        {category.category}
                      </span>
                      <span
                        className={cn(
                          'grid size-5 shrink-0 place-items-center rounded-full border text-[11px] font-bold',
                          checked ? 'border-primary bg-primary text-primary-foreground' : 'border-muted-foreground/30 text-transparent'
                        )}
                      >
                        ✓
                      </span>
                    </button>
                  )
                })}
              </div>
            )}
          </div>

          <DialogFooter>
            <Button variant='outline' onClick={closeDialog}>
              取消
            </Button>
            <Button
              onClick={handleSaveCategories}
              disabled={categoryLoading || categorySaving || categories.length === 0}
            >
              {categorySaving ? '保存中...' : '保存分类'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={isSourceDialog}
        onOpenChange={(val) => !val && closeDialog()}
      >
        <DialogContent className='max-h-[86vh] overflow-hidden p-0 sm:max-w-[820px]'>
          <DialogHeader className='border-b px-5 py-4 text-left'>
            <div className='flex items-start justify-between gap-4'>
              <div className='min-w-0'>
                <DialogTitle className='truncate text-lg'>
                  {source?.title || currentRow?.title || '查看帖子原文'}
                </DialogTitle>
                <DialogDescription className='mt-1 flex flex-wrap items-center gap-2'>
                  <span>原始 Markdown / 文本内容</span>
                  {source && (
                    <>
                      <Badge variant={source.articleStatus === 1 ? 'default' : 'secondary'}>
                        {source.articleStatus === 1 ? '发布' : '草稿'}
                      </Badge>
                      {source.processStatus === 1 && <Badge variant='destructive'>封禁</Badge>}
                    </>
                  )}
                </DialogDescription>
              </div>
            </div>
          </DialogHeader>

          <div className='grid min-h-0 gap-3 px-5 py-4'>
            {sourceLoading ? (
              <div className='rounded-lg border bg-muted/20 py-16 text-center text-sm text-muted-foreground'>
                正在加载原文...
              </div>
            ) : source ? (
              <>
                <div className='grid gap-2 rounded-lg border bg-muted/20 p-3 text-sm sm:grid-cols-[1fr_auto]'>
                  <div className='min-w-0 space-y-1'>
                    <div className='truncate font-medium text-foreground'>{source.title}</div>
                    <div className='flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground'>
                      <span>作者：{currentRow?.username || '未知用户'}</span>
                      <span>创建：{source.createdAt}</span>
                      <span>更新：{source.updatedAt}</span>
                    </div>
                    {source.description && (
                      <p className='line-clamp-2 text-xs leading-5 text-muted-foreground'>
                        {source.description}
                      </p>
                    )}
                  </div>
                  <div className='flex items-center gap-2 sm:justify-end'>
                    <Button variant='outline' size='sm' onClick={() => window.open(`/p/post/${source.id}`, '_blank')}>
                      打开前台
                    </Button>
                    <Button size='sm' onClick={copySource}>
                      复制原文
                    </Button>
                  </div>
                </div>

                <ScrollArea className='h-[52vh] rounded-lg border bg-slate-950 text-slate-50'>
                  <pre className='min-h-full whitespace-pre-wrap break-words p-4 font-mono text-xs leading-6 [overflow-wrap:anywhere]'>
                    {source.content || '暂无原文内容'}
                  </pre>
                </ScrollArea>
              </>
            ) : (
              <div className='rounded-lg border bg-muted/20 py-16 text-center text-sm text-muted-foreground'>
                未能加载原文内容
              </div>
            )}
          </div>

          <DialogFooter className='border-t px-5 py-3'>
            <Button variant='outline' onClick={closeDialog}>
              关闭
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <AlertDialog
        open={open !== null && open !== 'view' && open !== 'source' && open !== 'categories'}
        onOpenChange={(val) => !val && closeDialog()}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{getTitle()}</AlertDialogTitle>
            <AlertDialogDescription>
              {getDescription()}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={closeDialog}>
              取消
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleAction}
              className={open === 'reject' ? 'bg-destructive text-destructive-foreground hover:bg-destructive/90' : ''}
            >
              确认
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
